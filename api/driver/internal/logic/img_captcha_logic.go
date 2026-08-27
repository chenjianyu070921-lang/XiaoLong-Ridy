package logic

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math/big"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	ErrImgCaptchaInvalid = errors.New("图形验证码错误")
	ErrImgCaptchaExpired = errors.New("图形验证码已过期")
)

const imgCaptchaTTL = 5 * time.Minute

type ImgCaptchaLogic struct {
	ctx context.Context
}

type imgCaptchaEntry struct {
	Phone     string
	Code      string
	ExpiresAt time.Time
}

var globalImgCaptchaStore = &imgCaptchaStore{
	byUUID:  map[string]imgCaptchaEntry{},
	byPhone: map[string]string{},
}

type imgCaptchaStore struct {
	mu      sync.Mutex
	byUUID  map[string]imgCaptchaEntry
	byPhone map[string]string
}

func NewImgCaptchaLogic(ctx context.Context) *ImgCaptchaLogic {
	return &ImgCaptchaLogic{ctx: ctx}
}

func (l *ImgCaptchaLogic) Generate(phone string) (string, string, error) {
	phone = strings.TrimSpace(phone)
	if !validCaptchaPhone(phone) {
		return "", "", ErrInvalidParam
	}
	code, err := randomCaptchaCode(4)
	if err != nil {
		return "", "", err
	}
	id := uuid.NewString()
	img, err := renderCaptchaPNG(code)
	if err != nil {
		return "", "", err
	}
	globalImgCaptchaStore.save(phone, id, code, time.Now().Add(imgCaptchaTTL))
	return id, base64.StdEncoding.EncodeToString(img), nil
}

func (l *ImgCaptchaLogic) Verify(phone, id, input string) error {
	phone = strings.TrimSpace(phone)
	id = strings.TrimSpace(id)
	input = strings.TrimSpace(input)
	if !validCaptchaPhone(phone) || id == "" || input == "" {
		return ErrInvalidParam
	}
	return globalImgCaptchaStore.verify(phone, id, input, time.Now())
}

func (l *ImgCaptchaLogic) Invalidate(phone, id string) error {
	phone = strings.TrimSpace(phone)
	id = strings.TrimSpace(id)
	if !validCaptchaPhone(phone) || id == "" {
		return ErrInvalidParam
	}
	globalImgCaptchaStore.invalidate(phone, id)
	return nil
}

func (s *imgCaptchaStore) save(phone, id, code string, expiresAt time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if oldID := s.byPhone[phone]; oldID != "" {
		delete(s.byUUID, oldID)
	}
	s.byPhone[phone] = id
	s.byUUID[id] = imgCaptchaEntry{Phone: phone, Code: code, ExpiresAt: expiresAt}
}

func (s *imgCaptchaStore) verify(phone, id, input string, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.byUUID[id]
	if !ok || entry.Phone != phone {
		return ErrImgCaptchaInvalid
	}
	delete(s.byUUID, id)
	if s.byPhone[phone] == id {
		delete(s.byPhone, phone)
	}
	if now.After(entry.ExpiresAt) {
		return ErrImgCaptchaExpired
	}
	if !strings.EqualFold(entry.Code, input) {
		return ErrImgCaptchaInvalid
	}
	return nil
}

func (s *imgCaptchaStore) invalidate(phone, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.byUUID, id)
	if s.byPhone[phone] == id {
		delete(s.byPhone, phone)
	}
}

func validCaptchaPhone(phone string) bool {
	return regexp.MustCompile(`^1[3-9]\d{9}$`).MatchString(phone)
}

func randomCaptchaCode(length int) (string, error) {
	var b strings.Builder
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		b.WriteByte(byte('0' + n.Int64()))
	}
	return b.String(), nil
}

func renderCaptchaPNG(code string) ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, 118, 46))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.RGBA{R: 242, G: 247, B: 255, A: 255}}, image.Point{}, draw.Src)
	seed := captchaSeed(code)
	for i := 0; i < 8; i++ {
		x1 := int((seed + uint64(i*17)) % 118)
		y1 := int((seed/3 + uint64(i*11)) % 46)
		x2 := int((seed/7 + uint64(i*23)) % 118)
		y2 := int((seed/5 + uint64(i*19)) % 46)
		drawLine(img, x1, y1, x2, y2, color.RGBA{R: 89, G: 132, B: 214, A: 150})
	}
	for i, ch := range code {
		drawDigit(img, int(ch-'0'), 10+i*27, 9+(i%2)*2, color.RGBA{R: 20, G: 68, B: 150, A: 255})
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func captchaSeed(code string) uint64 {
	buf := make([]byte, 8)
	copy(buf, []byte(fmt.Sprintf("%8s", code)))
	return binary.LittleEndian.Uint64(buf)
}

var digitSegments = [10][7]bool{
	{true, true, true, true, true, true, false},
	{false, true, true, false, false, false, false},
	{true, true, false, true, true, false, true},
	{true, true, true, true, false, false, true},
	{false, true, true, false, false, true, true},
	{true, false, true, true, false, true, true},
	{true, false, true, true, true, true, true},
	{true, true, true, false, false, false, false},
	{true, true, true, true, true, true, true},
	{true, true, true, true, false, true, true},
}

func drawDigit(img *image.RGBA, digit, x, y int, c color.RGBA) {
	segments := digitSegments[digit]
	if segments[0] {
		fillRect(img, x+3, y, 15, 4, c)
	}
	if segments[1] {
		fillRect(img, x+18, y+3, 4, 12, c)
	}
	if segments[2] {
		fillRect(img, x+18, y+18, 4, 12, c)
	}
	if segments[3] {
		fillRect(img, x+3, y+30, 15, 4, c)
	}
	if segments[4] {
		fillRect(img, x, y+18, 4, 12, c)
	}
	if segments[5] {
		fillRect(img, x, y+3, 4, 12, c)
	}
	if segments[6] {
		fillRect(img, x+3, y+15, 15, 4, c)
	}
}

func fillRect(img *image.RGBA, x, y, w, h int, c color.RGBA) {
	draw.Draw(img, image.Rect(x, y, x+w, y+h), &image.Uniform{C: c}, image.Point{}, draw.Src)
}

func drawLine(img *image.RGBA, x0, y0, x1, y1 int, c color.RGBA) {
	dx := absInt(x1 - x0)
	dy := -absInt(y1 - y0)
	sx, sy := -1, -1
	if x0 < x1 {
		sx = 1
	}
	if y0 < y1 {
		sy = 1
	}
	err := dx + dy
	for {
		if image.Pt(x0, y0).In(img.Bounds()) {
			img.SetRGBA(x0, y0, c)
		}
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
