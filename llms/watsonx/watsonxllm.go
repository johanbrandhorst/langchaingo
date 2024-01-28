package watsonx

import (
	"context"

	wx "github.com/h0rv/go-watsonx/models"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
)

type LLM struct {
	CallbacksHandler callbacks.Handler
	client           *wx.Model
}

var _ llms.Model = (*LLM)(nil)

// Call implements the LLM interface.
func (o *LLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, o, prompt, options...)
}

// GenerateContent implements the Model interface.
func (o *LLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) { //nolint: lll, cyclop, whitespace

	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMGenerateContentStart(ctx, messages)
	}

	opts := &llms.CallOptions{}
	for _, opt := range options {
		opt(opts)
	}

	// Assume we get a single text message
	msg0 := messages[0]
	part := msg0.Parts[0]
	prompt := part.(llms.TextContent).Text
	result, err := o.client.GenerateText(
		prompt,
		toWatsonxOptions(opts)...,
	)
	if err != nil {
		if o.CallbacksHandler != nil {
			o.CallbacksHandler.HandleLLMError(ctx, err)
		}
		return nil, err
	}

	resp := &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content: result.Text,
			},
		},
	}
	return resp, nil
}

func New(opts ...wx.ModelOption) (*LLM, error) {
	c, err := wx.NewModel(opts...)
	if err != nil {
		return nil, err
	}

	return &LLM{
		client: c,
	}, nil
}

func toWatsonxOptions(opts *llms.CallOptions) []wx.GenerateOption {
	wxOpts := []wx.GenerateOption{
		wx.WithTopP(opts.TopP),
		wx.WithTopK(uint(opts.TopK)),
		// wx.WithDecodingMethod(decodingMethod), // not supported
		// wx.WithLengthPenalty(decayFactor, startIndex), // not supported
		wx.WithTemperature(opts.Temperature),
		wx.WithRandomSeed(uint(opts.Seed)),
		wx.WithRepetitionPenalty(uint(opts.RepetitionPenalty)),
		wx.WithMinNewTokens(uint(opts.MinLength)),
		wx.WithMaxNewTokens(uint(opts.MaxTokens)),
		wx.WithStopSequences(opts.StopWords),
		// wx.WithTimeLimit(timeLimit), // not supported
		// wx.WithTruncateInputTokens(truncateInputTokens), // not supported
		// wx.WithReturnOptions(inputText, generatedTokens, inputTokens, tokenLogProbs, tokenRanks, topNTokens), // not supported
	}

	return wxOpts
}