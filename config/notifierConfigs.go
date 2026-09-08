package config

type SamuraiNotifier struct {
	Ntfy    map[string]SamuraiNtfyNotifier
	Webhook map[string]SamuraiWebhookNotifier
}

type SamuraiNtfyNotifier struct {
	Server        string
	Topic         string
	Success       SamuariNtfyNotifierNotification
	Failure       SamuariNtfyNotifierNotification
	Update        SamuariNtfyNotifierNotification
	Authorization SamuraiNtfyNotifierAuthorization
}

type SamuraiNtfyNotifierAuthorization struct {
	Username string
	Password string
	Token    string
}

type SamuariNtfyNotifierNotification struct {
	Title    string
	Message  string
	Priority string
	Tags     string
}

type SamuraiWebhookNotifier struct {
	URL    string
	Method string
	Meta   map[string]string
}
