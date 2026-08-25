package task

import (
	"bamonC/model"
	"bamonC/request"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SyncBuddiesTask(ctx *gin.Context) {
	username := ctx.Param("username")
	err := BuddyService.DeleteBuddiesByUsername(username)
	if err != nil {
		log.Println(username, "::DeleteBuddiesByUsername: ", err)
		return
	}
	user, _ := UserService.GetByUsername(username)
	var buddies []*model.Buddy
	buddies, _ = request.GetBuddies(user)
	for _, buddy := range buddies {
		err := BuddyService.AddBuddy(buddy)
		if err != nil {
			log.Println(username, "::AddBuddy: ", err)
			return
		}
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "buddy sync successfully"})

}
