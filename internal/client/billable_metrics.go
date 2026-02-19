package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type BillableMetric struct {
	LagoID           string                 `json:"lago_id,omitempty"`
	Name             string                 `json:"name"`
	Code             string                 `json:"code"`
	Description      string                 `json:"description,omitempty"`
	AggregationType  string                 `json:"aggregation_type"`
	FieldName        string                 `json:"field_name,omitempty"`
	Expression       string                 `json:"expression,omitempty"`
	Recurring        bool                   `json:"recurring"`
	WeightedInterval string                 `json:"weighted_interval,omitempty"`
	Filters          []BillableMetricFilter `json:"filters,omitempty"`
	CreatedAt        string                 `json:"created_at,omitempty"`
	UpdatedAt        string                 `json:"updated_at,omitempty"`
}

type BillableMetricFilter struct {
	Key    string   `json:"key"`
	Values []string `json:"values"`
}

type CreateBillableMetricInput struct {
	Name             string                 `json:"name"`
	Code             string                 `json:"code"`
	Description      *string                `json:"description,omitempty"`
	AggregationType  string                 `json:"aggregation_type"`
	FieldName        *string                `json:"field_name,omitempty"`
	Expression       *string                `json:"expression,omitempty"`
	Recurring        *bool                  `json:"recurring,omitempty"`
	WeightedInterval *string                `json:"weighted_interval,omitempty"`
	Filters          []BillableMetricFilter `json:"filters,omitempty"`
}

type UpdateBillableMetricInput struct {
	Name             *string                `json:"name,omitempty"`
	Description      *string                `json:"description,omitempty"`
	AggregationType  *string                `json:"aggregation_type,omitempty"`
	FieldName        *string                `json:"field_name,omitempty"`
	Expression       *string                `json:"expression,omitempty"`
	Recurring        *bool                  `json:"recurring,omitempty"`
	WeightedInterval *string                `json:"weighted_interval,omitempty"`
	Filters          []BillableMetricFilter `json:"filters,omitempty"`
}

type billableMetricEnvelope struct {
	BillableMetric BillableMetric `json:"billable_metric"`
}

type createBillableMetricRequest struct {
	BillableMetric CreateBillableMetricInput `json:"billable_metric"`
}

type updateBillableMetricRequest struct {
	BillableMetric UpdateBillableMetricInput `json:"billable_metric"`
}

func (c *Client) CreateBillableMetric(ctx context.Context, input CreateBillableMetricInput) (*BillableMetric, error) {
	var envelope billableMetricEnvelope
	err := c.doRequest(ctx, http.MethodPost, "/api/v1/billable_metrics", createBillableMetricRequest{BillableMetric: input}, &envelope)
	if err != nil {
		return nil, fmt.Errorf("create billable metric: %w", err)
	}
	return &envelope.BillableMetric, nil
}

func (c *Client) GetBillableMetricByCode(ctx context.Context, code string) (*BillableMetric, error) {
	var envelope billableMetricEnvelope
	path := "/api/v1/billable_metrics/" + url.PathEscape(code)
	err := c.doRequest(ctx, http.MethodGet, path, nil, &envelope)
	if err != nil {
		return nil, fmt.Errorf("get billable metric %q: %w", code, err)
	}
	return &envelope.BillableMetric, nil
}

func (c *Client) UpdateBillableMetricByCode(ctx context.Context, code string, input UpdateBillableMetricInput) (*BillableMetric, error) {
	var envelope billableMetricEnvelope
	path := "/api/v1/billable_metrics/" + url.PathEscape(code)
	err := c.doRequest(ctx, http.MethodPut, path, updateBillableMetricRequest{BillableMetric: input}, &envelope)
	if err != nil {
		return nil, fmt.Errorf("update billable metric %q: %w", code, err)
	}
	return &envelope.BillableMetric, nil
}

func (c *Client) DeleteBillableMetricByCode(ctx context.Context, code string) error {
	path := "/api/v1/billable_metrics/" + url.PathEscape(code)
	err := c.doRequest(ctx, http.MethodDelete, path, nil, nil)
	if err != nil {
		return fmt.Errorf("delete billable metric %q: %w", code, err)
	}
	return nil
}
