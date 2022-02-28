package ancs_go

import (
	"errors"
	"github.com/godbus/dbus/v5"
	"github.com/muka/go-bluetooth/bluez/profile/device"
	"github.com/muka/go-bluetooth/bluez/profile/gatt"
)

//Ancs structure
type Ancs struct {
	dev              *device.Device1
	notification     *gatt.GattCharacteristic1
	control          *gatt.GattCharacteristic1
	data             *gatt.GattCharacteristic1
	NotificationChan chan notificationInfo
	DataChan         chan dataInfo
}
type notificationInfo struct {
	Original        []byte
	EventID         byte
	EventFlags      byte
	CategoryID      byte
	CategoryCount   byte
	NotificationUID []byte
}
type dataInfo struct {
	Original      []byte
	AppIdentifier string
	Title         string
	Subtitle      string
	Message       string
}

func NewAncs(objectPath dbus.ObjectPath) (*Ancs, error) {
	dev, e := device.NewDevice1(objectPath)
	if e != nil {
		return nil, e
	}
	notification, err := dev.GetCharByUUID("9FBF120D-6301-42D9-8C58-25E699A21DBD")
	if err != nil {
		return nil, err
	}
	control, err := dev.GetCharByUUID("69D1D8F3-45E1-49A8-9821-9BBDFDAAD9D9")
	if err != nil {
		return nil, err
	}
	data, err := dev.GetCharByUUID("22EAC6E9-24D6-4BB5-BE44-B36ACE7C7BFB")
	if err != nil {
		return nil, err
	}
	return &Ancs{
		dev:              dev,
		notification:     notification,
		control:          control,
		data:             data,
		NotificationChan: make(chan notificationInfo, 10),
		DataChan:         make(chan dataInfo, 10),
	}, nil
}
func (a *Ancs) StartNotify() error {
	notificationChannel, err := a.notification.WatchProperties()
	if err != nil {
		return err
	}
	dataChannel, err := a.data.WatchProperties()
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case notification := <-notificationChannel:
				if notification == nil {
					continue
				}
				if notification.Name != "Value" {
					continue
				}
				b1 := notification.Value.([]byte)
				a.NotificationChan <- notificationInfo{
					Original:        b1,
					EventID:         b1[0],
					EventFlags:      b1[1],
					CategoryID:      b1[2],
					CategoryCount:   b1[3],
					NotificationUID: b1[4:],
				}

			case data := <-dataChannel:
				if data == nil {
					continue
				}
				if data.Name != "Value" {
					continue
				}
				// [0 33 0 0 0 0 15 0 99 111 109 46 116 101 110 99 101 110 116 46 120 105 110 1 6 0 229 190 174 228 191
				//161 2 0 0 3 30 0 228 189 160 230 148 182 229 136 176 228 186 134 228 184 128 230 157 161 229 190 174
				//228 191 161 230 182 136 230 129 175]
				b1 := data.Value.([]byte)
				startN := byte(6)
				appIdentifierLen := b1[startN] + b1[startN+1]
				startN += 2
				//8,23
				appIdentifier := string(b1[startN : startN+appIdentifierLen])
				startN += appIdentifierLen + 1

				titleLen := b1[startN] + b1[startN+1]
				startN += 2
				title := string(b1[startN : startN+titleLen])
				startN += titleLen + 1

				subtitleLen := b1[startN] + b1[startN+1]
				startN += 2
				subtitle := string(b1[startN : startN+subtitleLen])
				startN += subtitleLen + 1
				messageLen := b1[startN] + b1[startN+1]
				startN += 2
				messageLenInt := int(messageLen)
				messageLenInt += int(startN)
				if messageLenInt > len(b1) {
					messageLenInt = len(b1)
				}
				message := string(b1[startN:messageLenInt])
				a.DataChan <- dataInfo{
					Original:      b1,
					AppIdentifier: appIdentifier,
					Title:         title,
					Subtitle:      subtitle,
					Message:       message,
				}
			}
		}
	}()
	err = a.notification.StartNotify()
	if err != nil {
		return err
	}
	err = a.data.StartNotify()
	if err != nil {
		return err
	}
	return nil
}

// SendToControl 发送信息至Control
func (a *Ancs) SendToControl(value []byte) error {
	options := make(map[string]interface{})
	err := a.control.WriteValue(value, options)
	if err != nil {
		return err
	}
	return nil
}

// GetNotificationAttributes 获取通知详情
func (a *Ancs) GetNotificationAttributes(notification notificationInfo) error {
	if notification.EventID == 0 {
		value := []byte{0}
		value = append(value, notification.NotificationUID...)
		value = append(value, 0, 1, 100, 0, 2, 100, 0, 3, 255, 0)
		err := a.SendToControl(value)
		if err != nil {
			return err
		}
		return nil
	}
	return errors.New("EventID cannot be 0")
}
