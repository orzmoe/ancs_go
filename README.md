# ancs_go

运行在树莓派上，用于获取iphone通知内容

需要树莓派有蓝牙功能

### 1. 将树莓派与手机配对

树莓派中依次运行

 ```
    bluetoothctl
    power on
    agent on
 ```

等待终端中出现手机的蓝牙名称 扫描到后执行

配对完成后记录下手机蓝牙的地址

```
    trust xx:xx:xx:xx:xx
    pair xx:xx:xx:xx:xx
```

### 2. 运行

下面代码中
> dev_xx_xx_xx_xx_xx_xx

修改为你的蓝牙id，蓝牙id中的冒号修改为下划线，如：蓝牙id为 DF:SD:QW:3S:SD 改为 dev_DF_SD_QW_3S_SD

```
func main() {
	ancs, err := NewAncs("/org/bluez/hci0/dev_xx_xx_xx_xx_xx_xx")
	if err != nil {
		panic(err)
	}
	err = ancs.StartNotify()
	if err != nil {
		fmt.Println(err)
	}

	for {
		select {
		case notification := <-ancs.NotificationChan:
			fmt.Println(notification.Original)
			err := ancs.GetNotificationAttributes(notification)
			if err != nil {
				fmt.Println(err.Error())
			}
		case data := <-ancs.DataChan:
			fmt.Println(data.Original)
			fmt.Printf("%s app:%s title:%s subtitle:%s message:%s \n", time.Now(), data.AppIdentifier, data.Title, data.Subtitle, data.Message)
		}
	}
}
```