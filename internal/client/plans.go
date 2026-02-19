package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type Plan struct {
	LagoID                  string                 `json:"lago_id,omitempty"`
	Name                    string                 `json:"name"`
	Code                    string                 `json:"code"`
	Interval                string                 `json:"interval"`
	Description             string                 `json:"description,omitempty"`
	AmountCents             int64                  `json:"amount_cents"`
	AmountCurrency          string                 `json:"amount_currency"`
	TrialPeriod             float64                `json:"trial_period,omitempty"`
	PayInAdvance            bool                   `json:"pay_in_advance"`
	BillChargesMonthly      bool                   `json:"bill_charges_monthly,omitempty"`
	BillFixedChargesMonthly bool                   `json:"bill_fixed_charges_monthly,omitempty"`
	InvoiceDisplayName      string                 `json:"invoice_display_name,omitempty"`
	TaxCodes                []string               `json:"tax_codes,omitempty"`
	Taxes                   []Tax                  `json:"taxes,omitempty"`
	Metadata                map[string]string      `json:"metadata,omitempty"`
	Charges                 []PlanCharge           `json:"charges,omitempty"`
	MinimumCommitment       *PlanMinimumCommitment `json:"minimum_commitment,omitempty"`
	FixedCharges            []PlanFixedCharge      `json:"fixed_charges,omitempty"`
	UsageThresholds         []PlanUsageThreshold   `json:"usage_thresholds,omitempty"`
	Entitlements            []PlanEntitlement      `json:"entitlements,omitempty"`
	CreatedAt               string                 `json:"created_at,omitempty"`
	UpdatedAt               string                 `json:"updated_at,omitempty"`
}

type PlanCharge struct {
	LagoBillableMetricID string             `json:"lago_billable_metric_id,omitempty"`
	BillableMetricID     string             `json:"billable_metric_id,omitempty"`
	ChargeModel          string             `json:"charge_model,omitempty"`
	Invoiceable          *bool              `json:"invoiceable,omitempty"`
	InvoiceDisplayName   *string            `json:"invoice_display_name,omitempty"`
	PayInAdvance         *bool              `json:"pay_in_advance,omitempty"`
	RegroupPaidFees      *bool              `json:"regroup_paid_fees,omitempty"`
	Prorated             *bool              `json:"prorated,omitempty"`
	MinAmountCents       *int64             `json:"min_amount_cents,omitempty"`
	Properties           json.RawMessage    `json:"properties,omitempty"`
	Filters              []PlanChargeFilter `json:"filters,omitempty"`
	TaxCodes             []string           `json:"tax_codes,omitempty"`
	Taxes                []Tax              `json:"taxes,omitempty"`
}

type PlanChargeFilter struct {
	LagoID             string              `json:"lago_id,omitempty"`
	ChargeCode         string              `json:"charge_code,omitempty"`
	InvoiceDisplayName *string             `json:"invoice_display_name,omitempty"`
	Properties         json.RawMessage     `json:"properties,omitempty"`
	Values             map[string][]string `json:"values,omitempty"`
}

type PlanMinimumCommitment struct {
	AmountCents        int64    `json:"amount_cents"`
	InvoiceDisplayName *string  `json:"invoice_display_name,omitempty"`
	TaxCodes           []string `json:"tax_codes,omitempty"`
	Taxes              []Tax    `json:"taxes,omitempty"`
}

type PlanFixedCharge struct {
	LagoAddOnID        string          `json:"lago_add_on_id,omitempty"`
	AddOnID            string          `json:"add_on_id,omitempty"`
	AddOnCode          string          `json:"add_on_code,omitempty"`
	ChargeModel        string          `json:"charge_model,omitempty"`
	InvoiceDisplayName *string         `json:"invoice_display_name,omitempty"`
	PayInAdvance       *bool           `json:"pay_in_advance,omitempty"`
	Prorated           *bool           `json:"prorated,omitempty"`
	Units              *int64          `json:"units,omitempty"`
	Properties         json.RawMessage `json:"properties,omitempty"`
	TaxCodes           []string        `json:"tax_codes,omitempty"`
	Taxes              []Tax           `json:"taxes,omitempty"`
}

type PlanUsageThreshold struct {
	AmountCents          *int64          `json:"amount_cents,omitempty"`
	ThresholdDisplayName *string         `json:"threshold_display_name,omitempty"`
	Recurring            *bool           `json:"recurring,omitempty"`
	Properties           json.RawMessage `json:"properties,omitempty"`
}

type PlanEntitlement struct {
	Code        string          `json:"code"`
	Name        *string         `json:"name,omitempty"`
	Description *string         `json:"description,omitempty"`
	Recurring   *bool           `json:"recurring,omitempty"`
	Privileges  json.RawMessage `json:"privileges,omitempty"`
}

type Tax struct {
	Code string `json:"code"`
}

type CreatePlanInput struct {
	Name                    string                 `json:"name"`
	Code                    string                 `json:"code"`
	Interval                string                 `json:"interval"`
	Description             *string                `json:"description,omitempty"`
	AmountCents             int64                  `json:"amount_cents"`
	AmountCurrency          string                 `json:"amount_currency"`
	TrialPeriod             *int64                 `json:"trial_period,omitempty"`
	PayInAdvance            *bool                  `json:"pay_in_advance,omitempty"`
	BillChargesMonthly      *bool                  `json:"bill_charges_monthly,omitempty"`
	BillFixedChargesMonthly *bool                  `json:"bill_fixed_charges_monthly,omitempty"`
	InvoiceDisplayName      *string                `json:"invoice_display_name,omitempty"`
	TaxCodes                []string               `json:"tax_codes,omitempty"`
	Metadata                map[string]string      `json:"metadata,omitempty"`
	Charges                 []PlanCharge           `json:"charges,omitempty"`
	MinimumCommitment       *PlanMinimumCommitment `json:"minimum_commitment,omitempty"`
	FixedCharges            []PlanFixedCharge      `json:"fixed_charges,omitempty"`
	UsageThresholds         []PlanUsageThreshold   `json:"usage_thresholds,omitempty"`
	Entitlements            []PlanEntitlement      `json:"entitlements,omitempty"`
}

type UpdatePlanInput struct {
	Name                    *string                `json:"name,omitempty"`
	Description             *string                `json:"description,omitempty"`
	Interval                *string                `json:"interval,omitempty"`
	AmountCents             *int64                 `json:"amount_cents,omitempty"`
	AmountCurrency          *string                `json:"amount_currency,omitempty"`
	TrialPeriod             *int64                 `json:"trial_period,omitempty"`
	PayInAdvance            *bool                  `json:"pay_in_advance,omitempty"`
	BillChargesMonthly      *bool                  `json:"bill_charges_monthly,omitempty"`
	BillFixedChargesMonthly *bool                  `json:"bill_fixed_charges_monthly,omitempty"`
	InvoiceDisplayName      *string                `json:"invoice_display_name,omitempty"`
	TaxCodes                []string               `json:"tax_codes,omitempty"`
	Metadata                map[string]string      `json:"metadata,omitempty"`
	Charges                 []PlanCharge           `json:"charges,omitempty"`
	MinimumCommitment       *PlanMinimumCommitment `json:"minimum_commitment,omitempty"`
	FixedCharges            []PlanFixedCharge      `json:"fixed_charges,omitempty"`
	UsageThresholds         []PlanUsageThreshold   `json:"usage_thresholds,omitempty"`
	Entitlements            []PlanEntitlement      `json:"entitlements,omitempty"`
}

type planEnvelope struct {
	Plan Plan `json:"plan"`
}

type createPlanRequest struct {
	Plan CreatePlanInput `json:"plan"`
}

type updatePlanRequest struct {
	Plan UpdatePlanInput `json:"plan"`
}

func (c *Client) CreatePlan(ctx context.Context, input CreatePlanInput) (*Plan, error) {
	var envelope planEnvelope
	err := c.doRequest(ctx, http.MethodPost, "/api/v1/plans", createPlanRequest{Plan: input}, &envelope)
	if err != nil {
		return nil, fmt.Errorf("create plan: %w", err)
	}
	return &envelope.Plan, nil
}

func (c *Client) GetPlanByCode(ctx context.Context, code string) (*Plan, error) {
	var envelope planEnvelope
	path := "/api/v1/plans/" + url.PathEscape(code)
	err := c.doRequest(ctx, http.MethodGet, path, nil, &envelope)
	if err != nil {
		return nil, fmt.Errorf("get plan %q: %w", code, err)
	}
	return &envelope.Plan, nil
}

func (c *Client) UpdatePlanByCode(ctx context.Context, code string, input UpdatePlanInput) (*Plan, error) {
	var envelope planEnvelope
	path := "/api/v1/plans/" + url.PathEscape(code)
	err := c.doRequest(ctx, http.MethodPut, path, updatePlanRequest{Plan: input}, &envelope)
	if err != nil {
		return nil, fmt.Errorf("update plan %q: %w", code, err)
	}
	return &envelope.Plan, nil
}

func (c *Client) DeletePlanByCode(ctx context.Context, code string) error {
	path := "/api/v1/plans/" + url.PathEscape(code)
	err := c.doRequest(ctx, http.MethodDelete, path, nil, nil)
	if err != nil {
		return fmt.Errorf("delete plan %q: %w", code, err)
	}
	return nil
}
