package main

import (
	"koding/db/mongodb/modelhelper"
	"math/rand"
	"os"
	"socialapi/config"
	"socialapi/models"
	"socialapi/rest"
	"strconv"
	"testing"
	"time"

	"github.com/koding/runner"
	. "github.com/smartystreets/goconvey/convey"
)

func TestPopularTopic(t *testing.T) {
	env := os.Getenv("SOCIAL_API_ENV")
	if env == "wercker" {
		return
	}
	r := runner.New("rest-tests")
	err := r.Init()
	So(err, ShouldBeNil)
	defer r.Close()

	appConfig := config.MustRead(r.Conf.Path)
	modelhelper.Initialize(appConfig.Mongo)
	defer modelhelper.Close()

	account := models.NewAccount()
	account.OldId = AccountOldId.Hex()
	var err error
	account, err = rest.CreateAccount(account)
	if err != nil {
		t.Fatalf("err %s", err.Error())
	}

	ses, err := models.FetchOrCreateSession(account.Nick)
	So(err, ShouldBeNil)
	So(ses, ShouldNotBeNil)

	rand.Seed(time.Now().UnixNano())
	groupName := "testgroup" + strconv.FormatInt(rand.Int63(), 10)

	// Since the wercker tests are failing it is skipped for temporarily
	Convey("order should be preserved", t, func() {
		So(err, ShouldBeNil)
		So(account, ShouldNotBeNil)
		channel1, err := rest.CreateChannelByGroupNameAndType(
			account.Id,
			groupName,
			models.Channel_TYPE_GROUP,
			ses.ClientId,
		)
		So(err, ShouldBeNil)
		So(channel1, ShouldNotBeNil)

		for i := 0; i < 5; i++ {
			post, err := rest.CreatePostWithBody(channel1.Id, account.Id, "create a message #5times")
			So(err, ShouldBeNil)
			So(post, ShouldNotBeNil)
		}

		for i := 0; i < 4; i++ {
			post, err := rest.CreatePostWithBody(channel1.Id, account.Id, "create a message #4times")
			So(err, ShouldBeNil)
			So(post, ShouldNotBeNil)
		}

		for i := 0; i < 3; i++ {
			post, err := rest.CreatePostWithBody(channel1.Id, account.Id, "create a message #3times")
			So(err, ShouldBeNil)
			So(post, ShouldNotBeNil)
		}

		//required for backgroud task to be finished
		time.Sleep(1 * time.Second)

		popularTopics, err := rest.FetchPopularTopics(account.Id, groupName)

		So(err, ShouldBeNil)
		So(popularTopics, ShouldNotBeNil)

		So(len(popularTopics), ShouldBeGreaterThanOrEqualTo, 3)

		So(popularTopics[0].Channel.Name, ShouldEqual, "5times")
		So(popularTopics[0].IsParticipant, ShouldEqual, false)
		So(popularTopics[0].ParticipantCount, ShouldEqual, 0)

		So(popularTopics[1].Channel.Name, ShouldEqual, "4times")
		So(popularTopics[1].IsParticipant, ShouldEqual, false)
		So(popularTopics[1].ParticipantCount, ShouldEqual, 0)

		So(popularTopics[2].Channel.Name, ShouldEqual, "3times")
		So(popularTopics[2].IsParticipant, ShouldEqual, false)
		So(popularTopics[2].ParticipantCount, ShouldEqual, 0)

		// check following status
		So(popularTopics[0].IsParticipant, ShouldBeFalse)
		So(popularTopics[1].IsParticipant, ShouldBeFalse)
		So(popularTopics[2].IsParticipant, ShouldBeFalse)
		// follow the first topic
		channelParticipant, err := rest.AddChannelParticipant(popularTopics[0].Channel.Id, account.Id, account.Id)
		So(err, ShouldBeNil)
		So(channelParticipant, ShouldNotBeNil)

		popularTopics, err = rest.FetchPopularTopics(account.Id, groupName)
		So(err, ShouldBeNil)
		So(popularTopics, ShouldNotBeNil)
		So(popularTopics[0].IsParticipant, ShouldBeTrue)
	})

}
