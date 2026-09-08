package driver

import (
	"context"
	"encoding/json"
	"strings"
)

// systemPrompt 与花小猪打车乘客评价司机的评分标准逐条对应，是打分的强制约束。
const systemPrompt = `你是网约车自动评价Agent，严格遵循花小猪打车乘客评价司机的评分标准，对已完成订单做自动打分。
# 花小猪评分规则
【星级范围】1-5星，5为最优，1最差。
评价参考信息来源：订单轨迹日志、乘客文字评价、接驾/行驶行为记录。

## 打分维度（4个核心维度）
1、接驾体验：司机接驾是否绕路、是否迟到、是否随意取消、是否提前点到达、是否多次不接乘客电话。
2、行驶体验：行车是否平稳、是否偏航、是否恶意绕路、是否遵守导航路线。
3、服务态度：司机言语、沟通，有无争吵、诱导乘客取消订单、言语冒犯。
4、车辆环境：车内整洁度，有无异味。

## 星级判定标准
5星优秀：无任何负面行为；接驾准时，路线正常，态度良好，车内干净。
4星良好：存在轻微小问题，不影响整体出行，无严重违规。
3星一般：有明显瑕疵，例如接驾慢、轻微绕路，但无恶意行为。
2星较差：出现明显违规，接驾严重超时、故意绕路、态度差。
1星极差：出现严重问题：诱导乘客取消、恶意拒接、辱骂、危险驾驶、严重绕路。

## 标签输出集合（只能从下面列表选，不能自己造标签）
正向标签：接驾准时｜路线合理｜服务态度好｜车内干净
负向标签：接驾迟到｜司机绕路｜司机态度差｜车内脏乱｜诱导取消｜不接电话｜提前点到达

# 强制约束
1.必须输出JSON，禁止输出多余自然语言，字段固定：
{"star":数字,"tags":["标签1","标签2"],"reason":"判定理由，简短说明依据，100字以内"}
2.没有对应标签就返回空数组[]；
3.优先乘客文本评价，再结合订单行为日志；
4.如果乘客无任何文字评价，只依据订单轨迹、业务行为做打分；
5.参考知识库参考资料；如果知识库为空，则只使用内置规则执行；
6.禁止编造不存在的订单行为。`

// buildUserPrompt 组装用户提示词：知识库片段 + 订单事实 JSON。
func buildUserPrompt(knowledgeContext string, order OrderInput) string {
	var sb strings.Builder
	sb.WriteString("【知识库参考资料】\n")
	if strings.TrimSpace(knowledgeContext) == "" {
		sb.WriteString("（空：知识库为空，仅使用内置评分规则执行）\n")
	} else {
		sb.WriteString(strings.TrimSpace(knowledgeContext))
		sb.WriteString("\n")
	}
	sb.WriteString("\n【订单输入信息】\norder_info：")
	raw, err := json.Marshal(order)
	if err != nil {
		raw = []byte("{}")
	}
	sb.Write(raw)
	return sb.String()
}

// retrieveKnowledge 读取知识库片段；检索失败视为知识库为空，不影响打分主链路。
func retrieveKnowledge(ctx context.Context, r KnowledgeRetriever, order OrderInput) string {
	if r == nil {
		return ""
	}
	text, err := r.Retrieve(ctx, order)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(text)
}
