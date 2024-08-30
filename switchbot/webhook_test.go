package switchbot_test

import (
	"bytes"
	"context"
	"encoding/json"
	switchbot2 "github.com/nasa9084/go-switchbot/v3/switchbot"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestWebhookSetup(t *testing.T) {
	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"statusCode":100,"body":{},"message":""}`))

			if r.Method != http.MethodPost {
				t.Fatalf("POST method is expected but %s", r.Method)
			}

			var got map[string]string
			if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
				t.Fatal(err)
			}

			want := map[string]string{
				"action":     "setupWebhook",
				"url":        "url1",
				"deviceList": "ALL",
			}

			if diff := cmp.Diff(want, got); diff != "" {
				t.Fatalf("event mismatch (-want +got):\n%s", diff)
			}
		}),
	)
	defer srv.Close()

	c := switchbot2.New("", "", switchbot2.WithEndpoint(srv.URL))

	if err := c.Webhook().Setup(context.Background(), "url1", "ALL"); err != nil {
		t.Fatal(err)
	}
}

func TestWebhookQuery(t *testing.T) {
	t.Run("queryUrl", func(t *testing.T) {
		srv := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{"statusCode":100,"body":{"urls":[url1]},"message":""}`))

				if r.Method != http.MethodPost {
					t.Fatalf("POST method is expected but %s", r.Method)
				}

				var got map[string]string
				if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
					t.Fatal(err)
				}

				want := map[string]string{
					"action": "queryUrl",
					"urls":   "",
				}

				if diff := cmp.Diff(want, got); diff != "" {
					t.Fatalf("event mismatch (-want +got):\n%s", diff)
				}
			}),
		)
		defer srv.Close()

		c := switchbot2.New("", "", switchbot2.WithEndpoint(srv.URL))

		if err := c.Webhook().Query(context.Background(), switchbot2.QueryURL, ""); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("queryDetails", func(t *testing.T) {
		srv := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{"statusCode":100,"body":[{"url":url1,"createTime":123456,"lastUpdateTime":123456,"deviceList":"ALL","enable":true}],"message":""}`))

				if r.Method != http.MethodPost {
					t.Fatalf("POST method is expected but %s", r.Method)
				}

				var got map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
					t.Fatal(err)
				}

				want := map[string]interface{}{
					"action": "queryDetails",
					"urls":   []interface{}{"url1"},
				}

				if diff := cmp.Diff(want, got); diff != "" {
					t.Fatalf("event mismatch (-want +got):\n%s", diff)
				}
			}),
		)
		defer srv.Close()

		c := switchbot2.New("", "", switchbot2.WithEndpoint(srv.URL))

		if err := c.Webhook().Query(context.Background(), switchbot2.QueryDetails, "url1"); err != nil {
			t.Fatal(err)
		}
	})
}

func TestWebhookUpdate(t *testing.T) {
	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"statusCode":100,"body":{},"message":""}`))

			if r.Method != http.MethodPost {
				t.Fatalf("POST method is expected but %s", r.Method)
			}

			var got map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
				t.Fatal(err)
			}

			want := map[string]interface{}{
				"action": "updateWebhook",
				"config": map[string]interface{}{
					"url":    "url1",
					"enable": true,
				},
			}

			if diff := cmp.Diff(want, got); diff != "" {
				t.Fatalf("event mismatch (-want +got):\n%s", diff)
			}
		}),
	)
	defer srv.Close()

	c := switchbot2.New("", "", switchbot2.WithEndpoint(srv.URL))

	if err := c.Webhook().Update(context.Background(), "url1", true); err != nil {
		t.Fatal(err)
	}
}

func TestWebhookDelete(t *testing.T) {
	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"statusCode":100,"body":{},"message":""}`))

			if r.Method != http.MethodPost {
				t.Fatalf("POST method is expected but %s", r.Method)
			}

			var got map[string]string
			if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
				t.Fatal(err)
			}

			want := map[string]string{
				"action": "deleteWebhook",
				"url":    "url1",
			}

			if diff := cmp.Diff(want, got); diff != "" {
				t.Fatalf("event mismatch (-want +got):\n%s", diff)
			}
		}),
	)
	defer srv.Close()

	c := switchbot2.New("", "", switchbot2.WithEndpoint(srv.URL))

	if err := c.Webhook().Delete(context.Background(), "url1"); err != nil {
		t.Fatal(err)
	}
}

func TestParseWebhook(t *testing.T) {
	sendWebhook := func(url, req string) {
		http.DefaultClient.Post(url, "application/json", bytes.NewBufferString(req))
	}

	t.Run("motion sensor", func(t *testing.T) {
		srv := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				event, err := switchbot2.ParseWebhookRequest(r)
				if err != nil {
					t.Fatal(err)
				}

				if got, ok := event.(*switchbot2.MotionSensorEvent); ok {
					want := switchbot2.MotionSensorEvent{
						EventType:    "changeReport",
						EventVersion: "1",
						Context: switchbot2.MotionSensorEventContext{
							DeviceType:     "WoPresence",
							DeviceMac:      "01:00:5e:90:10:00",
							DetectionState: "NOT_DETECTED",
							TimeOfSample:   123456789,
						},
					}

					if diff := cmp.Diff(want, *got); diff != "" {
						t.Fatalf("event mismatch (-want +got):\n%s", diff)
					}
				} else {
					t.Fatalf("given webhook event must be a motion sensor event but %T", event)
				}
			}),
		)
		defer srv.Close()

		sendWebhook(srv.URL, `{"eventType":"changeReport","eventVersion":"1","context": {"deviceType":"WoPresence","deviceMac":"01:00:5e:90:10:00","detectionState":"NOT_DETECTED","timeOfSample":123456789}}`)
	})

	t.Run("contact sensor", func(t *testing.T) {
		srv := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				event, err := switchbot2.ParseWebhookRequest(r)
				if err != nil {
					t.Fatal(err)
				}

				if got, ok := event.(*switchbot2.ContactSensorEvent); ok {
					want := switchbot2.ContactSensorEvent{
						EventType:    "changeReport",
						EventVersion: "1",
						Context: switchbot2.ContactSensorEventContext{
							DeviceType:     "WoContact",
							DeviceMac:      "01:00:5e:90:10:00",
							DetectionState: "NOT_DETECTED",
							DoorMode:       "OUT_DOOR",
							Brightness:     switchbot2.AmbientBrightnessDim,
							OpenState:      "open",
							TimeOfSample:   123456789,
						},
					}

					if diff := cmp.Diff(want, *got); diff != "" {
						t.Fatalf("event mismatch (-want +got):\n%s", diff)
					}
				} else {
					t.Fatalf("given webhook event must be a contact sensor event but %T", event)
				}
			}),
		)
		defer srv.Close()

		sendWebhook(srv.URL, `{"eventType":"changeReport","eventVersion":"1","context":{"deviceType":"WoContact","deviceMac":"01:00:5e:90:10:00","detectionState":"NOT_DETECTED","doorMode":"OUT_DOOR","brightness":"dim","openState":"open","timeOfSample":123456789}}`)
	})

	t.Run("meter", func(t *testing.T) {
		srv := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				event, err := switchbot2.ParseWebhookRequest(r)
				if err != nil {
					t.Fatal(err)
				}

				if got, ok := event.(*switchbot2.MeterEvent); ok {
					want := switchbot2.MeterEvent{
						EventType:    "changeReport",
						EventVersion: "1",
						Context: switchbot2.MeterEventContext{
							DeviceType:   "WoMeter",
							DeviceMac:    "01:00:5e:90:10:00",
							Temperature:  22.5,
							Scale:        "CELSIUS",
							Humidity:     31,
							TimeOfSample: 123456789,
						},
					}

					if diff := cmp.Diff(want, *got); diff != "" {
						t.Fatalf("event mismatch (-want +got):\n%s", diff)
					}
				} else {
					t.Fatalf("given webhook event must be a meter event but %T", event)
				}
			}),
		)
		defer srv.Close()

		sendWebhook(srv.URL, `{"eventType":"changeReport","eventVersion":"1","context":{"deviceType":"WoMeter","deviceMac":"01:00:5e:90:10:00","temperature":22.5,"scale":"CELSIUS","humidity":31,"timeOfSample":123456789}}`)
	})

	t.Run("meter plus", func(t *testing.T) {
		srv := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				event, err := switchbot2.ParseWebhookRequest(r)
				if err != nil {
					t.Fatal(err)
				}

				if got, ok := event.(*switchbot2.MeterPlusEvent); ok {
					want := switchbot2.MeterPlusEvent{
						EventType:    "changeReport",
						EventVersion: "1",
						Context: switchbot2.MeterPlusEventContext{
							DeviceType:   "WoMeterPlus",
							DeviceMac:    "01:00:5e:90:10:00",
							Temperature:  22.5,
							Scale:        "CELSIUS",
							Humidity:     31,
							TimeOfSample: 123456789,
						},
					}

					if diff := cmp.Diff(want, *got); diff != "" {
						t.Fatalf("event mismatch (-want +got):\n%s", diff)
					}
				} else {
					t.Fatalf("given webhook event must be a meter plus event but %T", event)
				}
			}),
		)
		defer srv.Close()

		// in the request body example the deviceType is Meter but I think it should be WoMeterPlus
		// https://github.com/OpenWonderLabs/SwitchBotAPI/blob/main/README-v1.0.md#meter-plus
		sendWebhook(srv.URL, `{"eventType":"changeReport","eventVersion":"1","context":{"deviceType":"WoMeterPlus","deviceMac":"01:00:5e:90:10:00","temperature":22.5,"scale":"CELSIUS","humidity":31,"timeOfSample":123456789}}`)
	})

	t.Run("lock", func(t *testing.T) {
		srv := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				event, err := switchbot2.ParseWebhookRequest(r)
				if err != nil {
					t.Fatal(err)
				}

				if got, ok := event.(*switchbot2.LockEvent); ok {
					want := switchbot2.LockEvent{
						EventType:    "changeReport",
						EventVersion: "1",
						Context: switchbot2.LockEventContext{
							DeviceType:   "WoLock",
							DeviceMac:    "01:00:5e:90:10:00",
							LockState:    "LOCKED",
							TimeOfSample: 123456789,
						},
					}

					if diff := cmp.Diff(want, *got); diff != "" {
						t.Fatalf("event mismatch (-want +got):\n%s", diff)
					}
				} else {
					t.Fatalf("given webhook event must be a lock event but %T", event)
				}
			}),
		)
		defer srv.Close()

		sendWebhook(srv.URL, `{"eventType":"changeReport","eventVersion":"1","context":{"deviceType":"WoLock","deviceMac":"01:00:5e:90:10:00","lockState":"LOCKED","timeOfSample":123456789}}`)
	})

	t.Run("indoor cam", func(t *testing.T) {
		srv := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				event, err := switchbot2.ParseWebhookRequest(r)
				if err != nil {
					t.Fatal(err)
				}

				if got, ok := event.(*switchbot2.IndoorCamEvent); ok {
					want := switchbot2.IndoorCamEvent{
						EventType:    "changeReport",
						EventVersion: "1",
						Context: switchbot2.IndoorCamEventContext{
							DeviceType:     "WoCamera",
							DeviceMac:      "01:00:5e:90:10:00",
							DetectionState: "DETECTED",
							TimeOfSample:   123456789,
						},
					}

					if diff := cmp.Diff(want, *got); diff != "" {
						t.Fatalf("event mismatch (-want +got):\n%s", diff)
					}
				} else {
					t.Fatalf("given webhook event must be a camera event but %T", event)
				}
			}),
		)
		defer srv.Close()

		sendWebhook(srv.URL, `{"eventType":"changeReport","eventVersion":"1","context":{"deviceType":"WoCamera","deviceMac":"01:00:5e:90:10:00","detectionState":"DETECTED","timeOfSample":123456789}}`)
	})

	t.Run("pan/tilt cam", func(t *testing.T) {
		srv := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				event, err := switchbot2.ParseWebhookRequest(r)
				if err != nil {
					t.Fatal(err)
				}

				if got, ok := event.(*switchbot2.PanTiltCamEvent); ok {
					want := switchbot2.PanTiltCamEvent{
						EventType:    "changeReport",
						EventVersion: "1",
						Context: switchbot2.PanTiltCamEventContext{
							DeviceType:     "WoPanTiltCam",
							DeviceMac:      "01:00:5e:90:10:00",
							DetectionState: "DETECTED",
							TimeOfSample:   123456789,
						},
					}

					if diff := cmp.Diff(want, *got); diff != "" {
						t.Fatalf("event mismatch (-want +got):\n%s", diff)
					}
				} else {
					t.Fatalf("given webhook event must be a pan/tilt camera event but %T", event)
				}
			}),
		)
		defer srv.Close()

		sendWebhook(srv.URL, `{"eventType":"changeReport","eventVersion":"1","context":{"deviceType":"WoPanTiltCam","deviceMac":"01:00:5e:90:10:00","detectionState":"DETECTED","timeOfSample":123456789}}`)
	})

	t.Run("color bulb", func(t *testing.T) {
		srv := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				event, err := switchbot2.ParseWebhookRequest(r)
				if err != nil {
					t.Fatal(err)
				}

				if got, ok := event.(*switchbot2.ColorBulbEvent); ok {
					want := switchbot2.ColorBulbEvent{
						EventType:    "changeReport",
						EventVersion: "1",
						Context: switchbot2.ColorBulbEventContext{
							DeviceType:       "WoBulb",
							DeviceMac:        "01:00:5e:90:10:00",
							PowerState:       switchbot2.PowerOn,
							Brightness:       10,
							Color:            "255:245:235",
							ColorTemperature: 3500,
							TimeOfSample:     123456789,
						},
					}

					if diff := cmp.Diff(want, *got); diff != "" {
						t.Fatalf("event mismatch (-want +got):\n%s", diff)
					}
				} else {
					t.Fatalf("given webhook event must be a color bulb event but %T", event)
				}
			}),
		)
		defer srv.Close()

		sendWebhook(srv.URL, `{"eventType":"changeReport","eventVersion":"1","context":{"deviceType":"WoBulb","deviceMac":"01:00:5e:90:10:00","powerState":"ON","brightness":10,"color":"255:245:235","colorTemperature":3500,"timeOfSample":123456789}}`)
	})

	t.Run("led strip light", func(t *testing.T) {
		srv := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				event, err := switchbot2.ParseWebhookRequest(r)
				if err != nil {
					t.Fatal(err)
				}

				if got, ok := event.(*switchbot2.StripLightEvent); ok {
					want := switchbot2.StripLightEvent{
						EventType:    "changeReport",
						EventVersion: "1",
						Context: switchbot2.StripLightEventContext{
							DeviceType:   "WoStrip",
							DeviceMac:    "01:00:5e:90:10:00",
							PowerState:   switchbot2.PowerOn,
							Brightness:   10,
							Color:        "255:245:235",
							TimeOfSample: 123456789,
						},
					}

					if diff := cmp.Diff(want, *got); diff != "" {
						t.Fatalf("event mismatch (-want +got):\n%s", diff)
					}
				} else {
					t.Fatalf("given webhook event must be a LED strip light event but %T", event)
				}
			}),
		)
		defer srv.Close()

		sendWebhook(srv.URL, `{"eventType":"changeReport","eventVersion":"1","context":{"deviceType":"WoStrip","deviceMac":"01:00:5e:90:10:00","powerState":"ON","brightness":10,"color":"255:245:235","timeOfSample":123456789}}`)
	})

	t.Run("plug mini (US)", func(t *testing.T) {
		srv := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				event, err := switchbot2.ParseWebhookRequest(r)
				if err != nil {
					t.Fatal(err)
				}

				if got, ok := event.(*switchbot2.PlugMiniUSEvent); ok {
					want := switchbot2.PlugMiniUSEvent{
						EventType:    "changeReport",
						EventVersion: "1",
						Context: switchbot2.PlugMiniUSEventContext{
							DeviceType:   "WoPlugUS",
							DeviceMac:    "01:00:5e:90:10:00",
							PowerState:   switchbot2.PowerOn,
							TimeOfSample: 123456789,
						},
					}

					if diff := cmp.Diff(want, *got); diff != "" {
						t.Fatalf("event mismatch (-want +got):\n%s", diff)
					}
				} else {
					t.Fatalf("given webhook event must be a plug mini (US) event but %T", event)
				}
			}),
		)
		defer srv.Close()

		sendWebhook(srv.URL, `{"eventType":"changeReport","eventVersion":"1","context":{"deviceType":"WoPlugUS","deviceMac":"01:00:5e:90:10:00","powerState":"ON","timeOfSample":123456789}}`)
	})

	t.Run("plug mini (JP)", func(t *testing.T) {
		srv := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				event, err := switchbot2.ParseWebhookRequest(r)
				if err != nil {
					t.Fatal(err)
				}

				if got, ok := event.(*switchbot2.PlugMiniJPEvent); ok {
					want := switchbot2.PlugMiniJPEvent{
						EventType:    "changeReport",
						EventVersion: "1",
						Context: switchbot2.PlugMiniJPEventContext{
							DeviceType:   "WoPlugJP",
							DeviceMac:    "01:00:5e:90:10:00",
							PowerState:   switchbot2.PowerOn,
							TimeOfSample: 123456789,
						},
					}

					if diff := cmp.Diff(want, *got); diff != "" {
						t.Fatalf("event mismatch (-want +got):\n%s", diff)
					}
				} else {
					t.Fatalf("given webhook event must be a plug mini (JP) event but %T", event)
				}
			}),
		)
		defer srv.Close()

		sendWebhook(srv.URL, `{"eventType":"changeReport","eventVersion":"1","context":{"deviceType":"WoPlugJP","deviceMac":"01:00:5e:90:10:00","powerState":"ON","timeOfSample":123456789}}`)
	})

	t.Run("Robot Vacuum Cleaner S1", func(t *testing.T) {
		srv := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				event, err := switchbot2.ParseWebhookRequest(r)
				if err != nil {
					t.Fatal(err)
				}

				if got, ok := event.(*switchbot2.SweeperEvent); ok {
					want := switchbot2.SweeperEvent{
						EventType:    "changeReport",
						EventVersion: "1",
						Context: switchbot2.SweeperEventContext{
							DeviceType:    "WoSweeper",
							DeviceMac:     "01:00:5e:90:10:00",
							WorkingStatus: switchbot2.CleanerStandBy,
							OnlineStatus:  switchbot2.CleanerOnline,
							Battery:       100,
							TimeOfSample:  123456789,
						},
					}

					if diff := cmp.Diff(want, *got); diff != "" {
						t.Fatalf("event mismatch (-want +got):\n%s", diff)
					}
				} else {
					t.Fatalf("given webhook event must be a sweeper event but %T", event)
				}
			}),
		)
		defer srv.Close()

		sendWebhook(srv.URL, `{"eventType":"changeReport","eventVersion":"1","context":{"deviceType":"WoSweeper","deviceMac":"01:00:5e:90:10:00","workingStatus":"StandBy","onlineStatus":"online","battery":100,"timeOfSample":123456789}}`)
	})

	t.Run("Robot Vacuum Cleaner S1 Plus", func(t *testing.T) {
		srv := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				event, err := switchbot2.ParseWebhookRequest(r)
				if err != nil {
					t.Fatal(err)
				}

				if got, ok := event.(*switchbot2.SweeperEvent); ok {
					want := switchbot2.SweeperEvent{
						EventType:    "changeReport",
						EventVersion: "1",
						Context: switchbot2.SweeperEventContext{
							DeviceType:    "WoSweeperPlus",
							DeviceMac:     "01:00:5e:90:10:00",
							WorkingStatus: switchbot2.CleanerStandBy,
							OnlineStatus:  switchbot2.CleanerOnline,
							Battery:       100,
							TimeOfSample:  123456789,
						},
					}

					if diff := cmp.Diff(want, *got); diff != "" {
						t.Fatalf("event mismatch (-want +got):\n%s", diff)
					}
				} else {
					t.Fatalf("given webhook event must be a sweeper plus event but %T", event)
				}
			}),
		)
		defer srv.Close()

		sendWebhook(srv.URL, `{"eventType":"changeReport","eventVersion":"1","context":{"deviceType":"WoSweeperPlus","deviceMac":"01:00:5e:90:10:00","workingStatus":"StandBy","onlineStatus":"online","battery":100,"timeOfSample":123456789}}`)
	})

	t.Run("Ceiling Light", func(t *testing.T) {
		srv := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				event, err := switchbot2.ParseWebhookRequest(r)
				if err != nil {
					t.Fatal(err)
				}

				if got, ok := event.(*switchbot2.CeilingEvent); ok {
					want := switchbot2.CeilingEvent{
						EventType:    "changeReport",
						EventVersion: "1",
						Context: switchbot2.CeilingEventContext{
							DeviceType:       "WoCeiling",
							DeviceMac:        "01:00:5e:90:10:00",
							PowerState:       switchbot2.PowerOn,
							Brightness:       10,
							ColorTemperature: 3500,
							TimeOfSample:     123456789,
						},
					}

					if diff := cmp.Diff(want, *got); diff != "" {
						t.Fatalf("event mismatch (-want +got):\n%s", diff)
					}
				} else {
					t.Fatalf("given webhook event must be a ceiling event but %T", event)
				}
			}),
		)
		defer srv.Close()

		sendWebhook(srv.URL, `{"eventType":"changeReport","eventVersion":"1","context":{"deviceType":"WoCeiling","deviceMac":"01:00:5e:90:10:00","powerState":"ON","brightness":10,"colorTemperature":3500,"timeOfSample":123456789}}`)
	})

	t.Run("Ceiling Light Pro", func(t *testing.T) {
		srv := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				event, err := switchbot2.ParseWebhookRequest(r)
				if err != nil {
					t.Fatal(err)
				}

				if got, ok := event.(*switchbot2.CeilingEvent); ok {
					want := switchbot2.CeilingEvent{
						EventType:    "changeReport",
						EventVersion: "1",
						Context: switchbot2.CeilingEventContext{
							DeviceType:       "WoCeilingPro",
							DeviceMac:        "01:00:5e:90:10:00",
							PowerState:       switchbot2.PowerOn,
							Brightness:       10,
							ColorTemperature: 3500,
							TimeOfSample:     123456789,
						},
					}

					if diff := cmp.Diff(want, *got); diff != "" {
						t.Fatalf("event mismatch (-want +got):\n%s", diff)
					}
				} else {
					t.Fatalf("given webhook event must be a ceiling event but %T", event)
				}
			}),
		)
		defer srv.Close()

		sendWebhook(srv.URL, `{"eventType":"changeReport","eventVersion":"1","context":{"deviceType":"WoCeilingPro","deviceMac":"01:00:5e:90:10:00","powerState":"ON","brightness":10,"colorTemperature":3500,"timeOfSample":123456789}}`)
	})

	t.Run("Keypad", func(t *testing.T) {
		t.Run("create a passcode", func(t *testing.T) {
			srv := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					event, err := switchbot2.ParseWebhookRequest(r)
					if err != nil {
						t.Fatal(err)
					}

					if got, ok := event.(*switchbot2.KeypadEvent); ok {
						want := switchbot2.KeypadEvent{
							EventType:    "changeReport",
							EventVersion: "1",
							Context: switchbot2.KeypadEventContext{
								DeviceType:   "WoKeypad",
								DeviceMac:    "01:00:5e:90:10:00",
								EventName:    "createKey",
								CommandID:    "CMD-1663558451952-01",
								Result:       "success",
								TimeOfSample: 123456789,
							},
						}

						if diff := cmp.Diff(want, *got); diff != "" {
							t.Fatalf("event mismatch (-want +got):\n%s", diff)
						}
					} else {
						t.Fatalf("given webhook event must be a keypad event but %T", event)
					}
				}),
			)
			defer srv.Close()

			sendWebhook(srv.URL, `{"eventType":"changeReport","eventVersion":"1","context":{"deviceType":"WoKeypad","deviceMac":"01:00:5e:90:10:00","eventName":"createKey","commandId":"CMD-1663558451952-01","result":"success","timeOfSample":123456789}}`)
		})
		t.Run("delete a passcode", func(t *testing.T) {
			srv := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					event, err := switchbot2.ParseWebhookRequest(r)
					if err != nil {
						t.Fatal(err)
					}

					if got, ok := event.(*switchbot2.KeypadEvent); ok {
						want := switchbot2.KeypadEvent{
							EventType:    "changeReport",
							EventVersion: "1",
							Context: switchbot2.KeypadEventContext{
								DeviceType:   "WoKeypad",
								DeviceMac:    "01:00:5e:90:10:00",
								EventName:    "deleteKey",
								CommandID:    "CMD-1663558451952-01",
								Result:       "success",
								TimeOfSample: 123456789,
							},
						}

						if diff := cmp.Diff(want, *got); diff != "" {
							t.Fatalf("event mismatch (-want +got):\n%s", diff)
						}
					} else {
						t.Fatalf("given webhook event must be a keypad event but %T", event)
					}
				}),
			)
			defer srv.Close()

			sendWebhook(srv.URL, `{"eventType":"changeReport","eventVersion":"1","context":{"deviceType":"WoKeypad","deviceMac":"01:00:5e:90:10:00","eventName":"deleteKey","commandId":"CMD-1663558451952-01","result":"success","timeOfSample":123456789}}`)
		})
	})

	t.Run("Keypad Touch", func(t *testing.T) {
		t.Run("create a passcode", func(t *testing.T) {
			srv := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					event, err := switchbot2.ParseWebhookRequest(r)
					if err != nil {
						t.Fatal(err)
					}

					if got, ok := event.(*switchbot2.KeypadEvent); ok {
						want := switchbot2.KeypadEvent{
							EventType:    "changeReport",
							EventVersion: "1",
							Context: switchbot2.KeypadEventContext{
								DeviceType:   "WoKeypadTouch",
								DeviceMac:    "01:00:5e:90:10:00",
								EventName:    "createKey",
								CommandID:    "CMD-1663558451952-01",
								Result:       "success",
								TimeOfSample: 123456789,
							},
						}

						if diff := cmp.Diff(want, *got); diff != "" {
							t.Fatalf("event mismatch (-want +got):\n%s", diff)
						}
					} else {
						t.Fatalf("given webhook event must be a keypad touch event but %T", event)
					}
				}),
			)
			defer srv.Close()

			sendWebhook(srv.URL, `{"eventType":"changeReport","eventVersion":"1","context":{"deviceType":"WoKeypadTouch","deviceMac":"01:00:5e:90:10:00","eventName":"createKey","commandId":"CMD-1663558451952-01","result":"success","timeOfSample":123456789}}`)
		})
		t.Run("delete a passcode", func(t *testing.T) {
			srv := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					event, err := switchbot2.ParseWebhookRequest(r)
					if err != nil {
						t.Fatal(err)
					}

					if got, ok := event.(*switchbot2.KeypadEvent); ok {
						want := switchbot2.KeypadEvent{
							EventType:    "changeReport",
							EventVersion: "1",
							Context: switchbot2.KeypadEventContext{
								DeviceType:   "WoKeypadTouch",
								DeviceMac:    "01:00:5e:90:10:00",
								EventName:    "deleteKey",
								CommandID:    "CMD-1663558451952-01",
								Result:       "success",
								TimeOfSample: 123456789,
							},
						}

						if diff := cmp.Diff(want, *got); diff != "" {
							t.Fatalf("event mismatch (-want +got):\n%s", diff)
						}
					} else {
						t.Fatalf("given webhook event must be a keypad touch event but %T", event)
					}
				}),
			)
			defer srv.Close()

			sendWebhook(srv.URL, `{"eventType":"changeReport","eventVersion":"1","context":{"deviceType":"WoKeypadTouch","deviceMac":"01:00:5e:90:10:00","eventName":"deleteKey","commandId":"CMD-1663558451952-01","result":"success","timeOfSample":123456789}}`)
		})
	})
}
