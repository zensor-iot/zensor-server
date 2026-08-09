package config

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"zensor-server/internal/infra/utils"

	"github.com/spf13/viper"
)

var loadConfigOnce sync.Once
var configInstance AppConfig

func LoadConfig() AppConfig {
	loadConfigOnce.Do(func() {
		viper.SetEnvPrefix("zensor_server")
		viper.AutomaticEnv()
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		viper.SetConfigName("server")
		viper.AddConfigPath("config")
		viper.AddConfigPath("/config")
		if err := viper.ReadInConfig(); err != nil {
			panic(fmt.Errorf("fatal error config file: %w", err))
		}
		configInstance = AppConfig{
			General: GeneralConfig{
				LogLevel: viper.GetString("general.log_level"),
			},
			HTTP: loadHTTPConfig(),
			mqtt: MqttConfig{
				Broker: viper.GetString("mqtt.broker"),
			},
			MQTTClient: MQTTClientConfig{
				Broker:   viper.GetString("mqtt_client.broker"),
				ClientID: viper.GetString("mqtt_client.client_id"),
				Username: viper.GetString("mqtt_client.username"),
				Password: viper.GetString("mqtt_client.password"),
			},
			Postgresql: PostgresqlConfig{
				DSN:          viper.GetString("database.dsn"),
				QueryTimeout: viper.GetDuration("database.query_timeout"),
			},
			Redis: RedisConfig{
				Addr:     viper.GetString("redis.addr"),
				Password: viper.GetString("redis.password"),
				DB:       viper.GetInt("redis.db"),
			},
			MailerSend: MailerSendConfig{
				APIKey:    viper.GetString("mailersend.api_key"),
				FromEmail: viper.GetString("mailersend.from_email"),
				FromName:  viper.GetString("mailersend.from_name"),
			},
			FCM: FCMConfig{
				ProjectID:          viper.GetString("fcm.project_id"),
				ServiceAccountPath: viper.GetString("fcm.service_account_path"),
			},
			WebPush: WebPushConfig{
				VAPIDPublicKey:  viper.GetString("notification.webpush.vapid_public_key"),
				VAPIDPrivateKey: viper.GetString("notification.webpush.vapid_private_key"),
				Subscriber:      viper.GetString("notification.webpush.subscriber"),
			},
			Auth:              loadAuthConfig(),
			Metrics:           loadMetricsConfig(),
			PushNotifications: loadPushNotificationsConfig(),
			Modules:           loadModulesConfig(),
			Victron:           loadVictronConfig(),
			ExecutionWorker: ExecutionWorkerConfig{
				TickerInterval: viper.GetDuration("execution_worker.ticker_interval"),
			},
		}
	})

	return configInstance
}

func loadHTTPConfig() HTTPConfig {
	port := viper.GetInt("http.port")
	if port == 0 {
		port = 3000
	}

	return HTTPConfig{
		Port: port,
	}
}

func loadAuthConfig() AuthConfig {
	sessionTTL := viper.GetDuration("auth.session_ttl")
	if sessionTTL == 0 {
		sessionTTL = 168 * time.Hour
	}

	mode := viper.GetString("auth.mode")
	if mode == "" {
		mode = AuthModeGoogle
	}

	return AuthConfig{
		Enabled:             viper.GetBool("auth.enabled"),
		Mode:                mode,
		SessionTTL:          sessionTTL,
		BootstrapAdminEmail: viper.GetString("auth.bootstrap_admin_email"),
		Google: GoogleOAuthConfig{
			ClientID:     viper.GetString("auth.google.client_id"),
			ClientSecret: viper.GetString("auth.google.client_secret"),
			RedirectURL:  viper.GetString("auth.google.redirect_url"),
		},
		Static: StaticAuthConfig{
			Username: viper.GetString("auth.static.username"),
			Password: viper.GetString("auth.static.password"),
		},
	}
}

func loadVictronConfig() VictronConfig {
	mqtt := VictronMQTTConfig{
		Broker:   viper.GetString("victron.mqtt.broker"),
		ClientID: viper.GetString("victron.mqtt.client_id"),
		Username: viper.GetString("victron.mqtt.username"),
		Password: viper.GetString("victron.mqtt.password"),
	}
	return VictronConfig{
		Enabled:  viper.GetBool("modules.victron.enabled"),
		PortalID: viper.GetString("victron.portal_id"),
		MQTT:     mqtt,
		CacheTTL: viper.GetDuration("victron.cache_ttl"),
	}
}

func loadMetricsConfig() []MetricWorkerConfig {
	metricsInterface := viper.Get("metrics")
	if metricsSlice, ok := metricsInterface.([]interface{}); ok {
		var metrics []MetricWorkerConfig
		for _, item := range metricsSlice {
			if metricMap, ok := item.(map[string]interface{}); ok {
				metric := MetricWorkerConfig{
					Name:              utils.ExtractStringValue(metricMap, "name"),
					Type:              utils.ExtractStringValue(metricMap, "type"),
					Topic:             utils.ExtractStringValue(metricMap, "topic"),
					EventType:         utils.ExtractStringValue(metricMap, "event_type"),
					ValuePropertyName: utils.ExtractStringValue(metricMap, "value_property_name"),
					CustomAttributes:  make(map[string]string),
				}
				if customAttributes, exists := metricMap["custom_attributes"]; exists {
					if attrMap, ok := customAttributes.(map[string]any); ok {
						for k, v := range attrMap {
							metric.CustomAttributes[k] = fmt.Sprintf("%v", v)
						}
					}
				}
				metrics = append(metrics, metric)
			}
		}
		return metrics
	}
	return []MetricWorkerConfig{}
}

func loadPushNotificationsConfig() []PushNotificationWorkerConfig {
	notificationsInterface := viper.Get("push_notifications")
	if notificationsSlice, ok := notificationsInterface.([]interface{}); ok {
		var notifications []PushNotificationWorkerConfig
		for _, item := range notificationsSlice {
			if notificationMap, ok := item.(map[string]interface{}); ok {
				notification := PushNotificationWorkerConfig{
					Name:             utils.ExtractStringValue(notificationMap, "name"),
					Topic:            utils.ExtractStringValue(notificationMap, "topic"),
					EventType:        utils.ExtractStringValue(notificationMap, "event_type"),
					TenantIDPath:     utils.ExtractStringValue(notificationMap, "tenant_id_path"),
					UserIDPath:       utils.ExtractStringValue(notificationMap, "user_id_path"),
					Title:            utils.ExtractStringValue(notificationMap, "title"),
					TitleTemplate:    utils.ExtractStringValue(notificationMap, "title_template"),
					Body:             utils.ExtractStringValue(notificationMap, "body"),
					BodyTemplate:     utils.ExtractStringValue(notificationMap, "body_template"),
					DeepLink:         utils.ExtractStringValue(notificationMap, "deeplink"),
					DeepLinkTemplate: utils.ExtractStringValue(notificationMap, "deeplink_template"),
				}
				notifications = append(notifications, notification)
			}
		}
		return notifications
	}
	return []PushNotificationWorkerConfig{}
}

func loadModulesConfig() ModulesConfig {
	return ModulesConfig{
		Permaculture: ModuleConfig{
			Enabled: viper.GetBool("modules.permaculture.enabled"),
		},
		Maintenance: ModuleConfig{
			Enabled: viper.GetBool("modules.maintenance.enabled"),
		},
		Victron: ModuleConfig{
			Enabled: viper.GetBool("modules.victron.enabled"),
		},
	}
}

type AppConfig struct {
	General           GeneralConfig
	HTTP              HTTPConfig
	Auth              AuthConfig
	mqtt              MqttConfig
	MQTTClient        MQTTClientConfig
	Victron           VictronConfig
	Postgresql        PostgresqlConfig
	Redis             RedisConfig
	MailerSend        MailerSendConfig
	FCM               FCMConfig
	WebPush           WebPushConfig
	Metrics           MetricsConfig
	PushNotifications PushNotificationsConfig
	Modules           ModulesConfig
	ExecutionWorker   ExecutionWorkerConfig
}

type GeneralConfig struct {
	LogLevel string
}

type HTTPConfig struct {
	Port int
}

// AuthModeGoogle authenticates users via Google OAuth against the allowlist.
// AuthModeStatic authenticates a single hardcoded admin user; local dev only.
const (
	AuthModeGoogle = "google"
	AuthModeStatic = "static"
)

type AuthConfig struct {
	Enabled             bool
	Mode                string
	Google              GoogleOAuthConfig
	Static              StaticAuthConfig
	SessionTTL          time.Duration
	BootstrapAdminEmail string
}

type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type StaticAuthConfig struct {
	Username string
	Password string
}

type MqttConfig struct {
	Broker string
}

type MQTTClientConfig struct {
	Broker   string
	ClientID string
	Username string
	Password string
}

type PostgresqlConfig struct {
	DSN          string
	QueryTimeout time.Duration
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type MailerSendConfig struct {
	APIKey    string
	FromEmail string
	FromName  string
}

type FCMConfig struct {
	ProjectID          string
	ServiceAccountPath string
}

type WebPushConfig struct {
	VAPIDPublicKey  string
	VAPIDPrivateKey string
	Subscriber      string
}

type MetricsConfig []MetricWorkerConfig

type PushNotificationsConfig []PushNotificationWorkerConfig

type PushNotificationWorkerConfig struct {
	Name             string
	Topic            string
	EventType        string
	TenantIDPath     string
	UserIDPath       string
	Title            string
	TitleTemplate    string
	Body             string
	BodyTemplate     string
	DeepLink         string
	DeepLinkTemplate string
}

type MetricWorkerConfig struct {
	Name              string
	Type              string
	Topic             string
	EventType         string
	ValuePropertyName string
	CustomAttributes  map[string]string
}

type ModulesConfig struct {
	Permaculture ModuleConfig
	Maintenance  ModuleConfig
	Victron      ModuleConfig
}

type VictronConfig struct {
	Enabled  bool
	PortalID string
	MQTT     VictronMQTTConfig
	CacheTTL time.Duration
}

type VictronMQTTConfig struct {
	Broker   string
	ClientID string
	Username string
	Password string
}

type ModuleConfig struct {
	Enabled bool
}

type ExecutionWorkerConfig struct {
	TickerInterval time.Duration
}
