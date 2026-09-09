package react

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// RegisteredTool is the small executable tool contract used by this skeleton.
// It also implements Eino's BaseTool and InvokableTool interfaces.
type RegisteredTool struct {
	InfoData *schema.ToolInfo
	Run      func(context.Context, string) (string, error)
}

func (t *RegisteredTool) Info(context.Context) (*schema.ToolInfo, error) {
	return t.InfoData, nil
}

func (t *RegisteredTool) InvokableRun(ctx context.Context, argumentsInJSON string, _ ...tool.Option) (string, error) {
	return t.Run(ctx, argumentsInJSON)
}

// ToolRegistry stores the tools exposed to the model and executed by the graph.
type ToolRegistry struct {
	tools map[string]*RegisteredTool
}

func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{tools: make(map[string]*RegisteredTool)}
}

func (r *ToolRegistry) Register(t *RegisteredTool) error {
	if t == nil || t.InfoData == nil || t.InfoData.Name == "" || t.Run == nil {
		return fmt.Errorf("invalid tool registration")
	}
	if _, exists := r.tools[t.InfoData.Name]; exists {
		return fmt.Errorf("tool %q already registered", t.InfoData.Name)
	}
	r.tools[t.InfoData.Name] = t
	return nil
}

func (r *ToolRegistry) Infos() []*schema.ToolInfo {
	infos := make([]*schema.ToolInfo, 0, len(r.tools))
	for _, t := range r.tools {
		infos = append(infos, t.InfoData)
	}
	return infos
}

func (r *ToolRegistry) Execute(ctx context.Context, call schema.ToolCall) (string, error) {
	t, ok := r.tools[call.Function.Name]
	if !ok {
		return "", fmt.Errorf("tool %q is not registered", call.Function.Name)
	}
	return t.InvokableRun(ctx, call.Function.Arguments)
}

func NewDefaultToolRegistry() (*ToolRegistry, error) {
	registry := NewToolRegistry()
	if err := registry.Register(&RegisteredTool{
		InfoData: &schema.ToolInfo{
			Name: "get_product_price",
			Desc: "查询商品价格。product_id 使用商品编号，例如 1001。",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"product_id": {Type: schema.String, Desc: "商品编号", Required: true},
			}),
		},
		Run: func(_ context.Context, input string) (string, error) {
			var args struct {
				ProductID string `json:"product_id"`
			}
			if err := json.Unmarshal([]byte(input), &args); err != nil {
				return "", fmt.Errorf("invalid product arguments: %w", err)
			}
			if args.ProductID != "1001" {
				return fmt.Sprintf("商品%s暂无 mock 价格", args.ProductID), nil
			}
			return "商品1001价格为299元", nil
		},
	}); err != nil {
		return nil, err
	}
	if err := registry.Register(&RegisteredTool{
		InfoData: &schema.ToolInfo{
			Name: "get_current_time",
			Desc: "获取当前服务器时间。",
		},
		Run: func(_ context.Context, _ string) (string, error) {
			return time.Now().Format(time.RFC3339), nil
		},
	}); err != nil {
		return nil, err
	}
	return registry, nil
}
