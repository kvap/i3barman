package battery

import (
	"fmt"
	"time"

	"github.com/godbus/dbus"
)

type Manager struct {
	conn *dbus.Conn
}

func NewManager() (*Manager, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, err
	}

	return &Manager{conn}, nil
}

func (bm *Manager) getDisplayDeviceProperties() map[string]dbus.Variant {
	obj := bm.conn.Object("org.freedesktop.UPower", "/org/freedesktop/UPower/devices/DisplayDevice")

	var props map[string]dbus.Variant

	err := obj.Call("org.freedesktop.DBus.Properties.GetAll", 0, "org.freedesktop.UPower.Device").Store(&props)
	if err != nil {
		panic(err)
	}

	return props
}

type Status struct {
	Percentage float64
	TimeToEmpty time.Duration
	TimeToFull time.Duration
	Rate float64
	Charging bool
	Discharging bool
}

func (bm *Manager) GetStatus() (status *Status, err error) {
	defer func() {
		if r := recover(); r != nil {
			status = nil
			err = fmt.Errorf("recovered in GetStatus: %v", r)
		}
	}()

	props := bm.getDisplayDeviceProperties()

	return &Status{
		Percentage: props["Percentage"].Value().(float64),
		TimeToEmpty: time.Second * time.Duration(props["TimeToEmpty"].Value().(int64)),
		TimeToFull: time.Second * time.Duration(props["TimeToFull"].Value().(int64)),
		Rate: props["EnergyRate"].Value().(float64),
		Charging: props["State"].Value().(uint32) == 1,
		Discharging: props["State"].Value().(uint32) == 2,
	}, nil
}
