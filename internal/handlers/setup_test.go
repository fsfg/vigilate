package handlers

import (
	"context"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/pusher/pusher-http-go"
	"github.com/robfig/cron/v3"

	"gitlab.com/fsfg/vigilate/internal/channeldata"
	"gitlab.com/fsfg/vigilate/internal/config"
	"gitlab.com/fsfg/vigilate/internal/driver"
	"gitlab.com/fsfg/vigilate/internal/helpers"
	"gitlab.com/fsfg/vigilate/internal/repository/dbrepo"
)

var testSession *scs.SessionManager

func TestMain(m *testing.M) {

	testSession = scs.New()
	testSession.Lifetime = 24 * time.Hour
	testSession.Cookie.Persist = true
	testSession.Cookie.SameSite = http.SameSiteLaxMode
	testSession.Cookie.Secure = false

	mailQueue := make(chan channeldata.MailJob, 5)

	// define application configuration
	a := config.AppConfig{
		DB:           &driver.DB{},
		Session:      testSession,
		InProduction: false,
		Domain:       "localhost",
		MailQueue:    mailQueue,
	}

	app = &a

	preferenceMap := map[string]string{}
	app.PreferenceMap = preferenceMap

	// create pusher client
	dws := dummyWS{
		AppID:  "1",
		Secret: "123abc",
		Key:    "abc123",
		Secure: false,
		Host:   "localhost:4001",
	}
	app.WsClient = &dws

	monitorMap := map[int]cron.EntryID{}
	app.MonitorMap = monitorMap

	localZone, _ := time.LoadLocation("Local")
	scheduler := cron.New(
		cron.WithLocation(localZone),
		cron.WithChain(
			cron.DelayIfStillRunning(cron.DefaultLogger),
			cron.Recover(cron.DefaultLogger),
		),
	)
	app.Scheduler = scheduler

	repo := NewTestHandlers(app)
	NewHandlers(repo, app)

	helpers.NewHelpers(app)

	helpers.SetViews("./../../views")

	os.Exit(m.Run())
}

// gets the context with session added
func getCtx(req *http.Request) context.Context {
	ctx, err := testSession.Load(
		req.Context(),
		req.Header.Get("X-Session"),
	)
	if err != nil {
		log.Println(err)
	}
	return ctx
}

// NewTestHandlers creates a new test repository
func NewTestHandlers(a *config.AppConfig) *DBRepo {
	return &DBRepo{
		App: a,
		DB:  dbrepo.NewTestingRepo(a),
	}
}

// dummyWS is a pusher.Client implementation
type dummyWS struct {
	AppID                        string
	Key                          string
	Secret                       string
	Host                         string
	Secure                       bool
	Cluster                      string
	HTTPClient                   *http.Client
	EncryptionMasterKey          string
	EncryptionMasterKeyBase64    string
	validatedEncryptionMasterKey *[]byte
}

func (c *dummyWS) Trigger(channel string, eventName string, data any) error {
	return nil
}

func (c *dummyWS) TriggerMulti(channels []string, eventName string, data any) error {
	return nil
}

func (c *dummyWS) TriggerExclusive(channel string, eventName string, data any, socketID string) error {
	return nil
}

func (c *dummyWS) TriggerMultiExclusive(channels []string, eventName string, data any, socketID string) error {
	return nil
}

func (c *dummyWS) TriggerBatch(batch []pusher.Event) error {
	return nil
}

func (c *dummyWS) Channels(additionalQueries map[string]string) (*pusher.ChannelsList, error) {
	return &pusher.ChannelsList{}, nil
}

func (c *dummyWS) Channel(name string, additionalQueries map[string]string) (*pusher.Channel, error) {
	return &pusher.Channel{}, nil
}

func (c *dummyWS) GetChannelUsers(name string) (*pusher.Users, error) {
	return &pusher.Users{}, nil
}

func (c *dummyWS) AuthenticatePrivateChannel(params []byte) (response []byte, err error) {
	response, err = []byte{}, nil
	return
}

func (c *dummyWS) AuthenticatePresenceChannel(params []byte, member pusher.MemberData) (response []byte, err error) {
	jsonStr := `{
		"auth":"abc123:746cd5a384b9876abb62a4ed3d0da3aebfa56aa91d9e2b773a40bf64b4485add",
		"channel_data":"{\"user_id\":\"2\",\"user_info\":{\"id\":\"2\",\"name\":\"Agent\"}}"
	}`
	response, err = []byte(jsonStr), nil
	return
}

func (c *dummyWS) Webhook(header http.Header, body []byte) (*pusher.Webhook, error) {
	return &pusher.Webhook{}, nil
}
