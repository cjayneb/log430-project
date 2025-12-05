package core

import (
	"brokerx/notification-service/models"
	"brokerx/notification-service/ports"
	"brokerx/notification-service/util"
	"context"
	"fmt"

	"github.com/resend/resend-go/v2"
)

type NotificationService interface {
	SendNotification(ctx context.Context, event models.Order) error
}

type NotificationServiceImpl struct {
	NotificationRepo  ports.NotificationRepository
	UserRepository    ports.UserRepository
	UserServiceClient ports.UserService
	Producer          ports.EventProducer
}

// TODO: implement quasi strategy pattern

func (n *NotificationServiceImpl) SendNotification(ctx context.Context, order models.Order) error {
	log := util.FromContext(ctx)
	log.Info("Sending email for Order", "orderId", order.ID)

	apiKey := "re_beymtupC_FCP32S5Gi1vWoHw3YPwvf4jm"

	client := resend.NewClient(apiKey)

	preferences, err := n.NotificationRepo.FindByUserId(ctx, order.UserID)
	if err != nil {
		log.Error("error fetching user notification preferences", "error", err)
		return err
	}

	info, err := n.UserRepository.GetContactInfo(ctx, order.UserID)
	if err != nil {
		log.Error("error when fetching user contact info", "error", err)
		return err
	}
	var empty = models.UserContactInfo{}
	if empty == info {
		log.Info("no contact info found for user", "userId", order.UserID)
		userInfo, err := n.UserServiceClient.GetUserContactInfo(ctx, order.UserID)
		if err != nil {
			log.Error("error when fetching user contact info from user service", "error", err)
			return err
		}
		info = userInfo

		err = n.UserRepository.SetContactInfo(ctx, userInfo)
		if err != nil {
			log.Warn("error when caching user contact info", "error", err)
		}
	}

	if preferences.Email {
		params := &resend.SendEmailRequest{
			From:    "brokerx@jcbenoit.ca",
			To:      []string{info.Email},
			Subject: getEmailSubject(order),
			Html:    getEmailBody(order),
		}

		_, err := client.Emails.Send(params)
		if err != nil {
			log.Error("error when sending email", "error", err)
			return err
		}
	}

	return nil
}

func getEmailBody(order models.Order) string {
	msg := "Your order #%d has been filled."
	if order.Status == "canceled" {
		msg = "Your order #%d has been canceled."
	}
	if order.Status == "partially_filled" {
		msg = "Your order #%d has been partially filled and the remaining quantity has been canceled."
	}
	return fmt.Sprintf(msg, order.ID)
}

func getEmailSubject(order models.Order) string {
	msg := "Order #%d Filled"
	if order.Status == "canceled" {
		msg = "Order #%d Canceled"
	}
	if order.Status == "partially_filled" {
		msg = "Order #%d Partially Filled"
	}
	return fmt.Sprintf(msg, order.ID)
}

var _ NotificationService = (*NotificationServiceImpl)(nil) // Ensure interface is implemented at compile time
