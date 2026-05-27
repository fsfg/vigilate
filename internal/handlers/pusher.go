package handlers

import (
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/pusher/pusher-http-go"
)

// PusherAuth authenticates the user to our pusher server
func (repo *DBRepo) PusherAuth(w http.ResponseWriter, r *http.Request) {
	userID := repo.App.Session.GetInt(r.Context(), "userID")

	u, _ := repo.DB.GetUserByID(userID)

	params, _ := io.ReadAll(r.Body)

	presenceData := pusher.MemberData{
		UserID: strconv.Itoa(userID),
		UserInfo: map[string]string{
			"name": u.FirstName,
			"id":   strconv.Itoa(userID),
		},
	}

	response, err := app.WsClient.AuthenticatePresenceChannel(params, presenceData)
	if err != nil {
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	_, _ = w.Write(response)
}

// TestPusher just tests pusher - delete this before going into production
func (repo *DBRepo) TestPusher(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"message": "hello world",
	}

	if err := repo.App.WsClient.
		Trigger("public-channel", "test-event", data); err != nil {
		log.Println(err)
	}

}
