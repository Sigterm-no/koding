package webhook

import (
	"math/rand"
	"socialapi/config"
	"socialapi/models"
	"testing"
	"time"

	"github.com/koding/logging"
	"github.com/koding/runner"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSendMessage(t *testing.T) {

	r := runner.New("test")
	if err := r.Init(); err != nil {
		t.Fatalf("something went wrong: %s", err)
	}
	config.MustRead(r.Conf.Path)
	r.Log.SetLevel(logging.CRITICAL)

	defer r.Close()

	Convey("while testing bot", t, func() {

		bot, err := NewBot()
		So(err, ShouldBeNil)

		rand.Seed(time.Now().UTC().UnixNano())
		groupName := models.RandomName()

		channel := models.CreateTypedGroupedChannelWithTest(bot.account.Id, models.Channel_TYPE_TOPIC, groupName)

		Convey("bot should be able to create message", func() {
			message := &Message{}
			message.Body = "testmessage"
			message.ChannelId = channel.Id
			err := bot.SendMessage(message)
			So(err, ShouldBeNil)

			m, err := channel.FetchLastMessage()
			So(err, ShouldBeNil)
			So(m, ShouldNotBeNil)
			So(m.Body, ShouldEqual, message.Body)
			So(m.InitialChannelId, ShouldEqual, message.ChannelId)
			So(m.AccountId, ShouldEqual, bot.account.Id)

		})
	})
}
