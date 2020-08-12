package fcm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
)

const (
	// MinTTL is the minimum value used for the time to live field of FCM service.
	MinTTL = 0

	// MaxTTL is the default maximum time the notification will live to be sent to our users.
	MaxTTL = 86400 * 7 // 1 week

	// Endpoint is the send notification URI for FCM.
	Endpoint = "https://fcm.googleapis.com/fcm/send"

	// FCMAuthKey env name
	FCMAuthKey = "FCM_AUTH_KEY"

	// FCMAuthKey command placeholder
	FCMAuthKeyPlaceholder = "YOUR_FCM_AUTH_KEY"
)

var (
	ErrInvalidNotification  = errors.New("notification is invalid")
	ErrFcmAuthKeyNotDefined = errors.New("FCM auth key not defined")
)

// Notification ...
type Notification struct {
	Topic string      `json:"topic"`
	Data  interface{} `json:"data"`
}

// NotificationPostBody is used to post notification to FCM servers
type NotificationPostBody struct {
	To         string      `json:"to"`
	Data       interface{} `json:"data,omitempty"`
	TimeToLive int         `json:"time-to-live,omitempty"`
}

// SendNotification ...
func SendNotification(notification *Notification) (bool, error) {
	fcmAuthKey := os.Getenv("FCM_AUTH_KEY")
	if fcmAuthKey == "" {
		return false, ErrFcmAuthKeyNotDefined
	}
	if notification == nil || notification.Data == nil {
		return false, ErrInvalidNotification
	}

	// Let's use half a week, which will make possible to users receive this notification
	// before next Monday (if this code run in a Thursday)
	ttl := MaxTTL / 2

	notificationBody := NotificationPostBody{
		To:         "/topics/" + notification.Topic,
		Data:       notification.Data,
		TimeToLive: ttl,
	}

	buff, err := json.Marshal(notificationBody)
	if err != nil {
		return false, err
	}

	req, err := http.NewRequest("POST", Endpoint, bytes.NewBuffer(buff))
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", "key="+fcmAuthKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	success := resp.StatusCode >= 200 && resp.StatusCode < 300
	return success, nil
}

// NewNotification ...
func NewNotification(topic string, data interface{}) *Notification {
	if data == nil {
		return nil
	}

	return &Notification{
		Topic: topic,
		Data:  data,
	}
}

// NewTheaterTopic ...
func NewTheaterTopic(ID int) string {
	return fmt.Sprintf("theater_%d", ID)
}
