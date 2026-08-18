package billing

// Provider is the checkout/webhook adapter. Swap implementations when a vendor is chosen.
type Provider interface {
	CheckoutURL(userID, plan string) (string, error)
	ApplyWebhook(payload []byte) (Grant, error)
}

type Grant struct {
	UserID  string
	Credits int
	Plan    string
}

type Fake struct {
	Site string
}

func (f Fake) CheckoutURL(userID, plan string) (string, error) {
	site := f.Site
	if site == "" {
		site = "https://loremetry.com"
	}
	if plan == "" {
		plan = "writer"
	}
	return site + "/api/billing/fake-checkout?user=" + userID + "&plan=" + plan, nil
}

func (f Fake) ApplyWebhook(payload []byte) (Grant, error) {
	return Grant{}, nil
}
