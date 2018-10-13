package network

import (
	"fmt"

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

func (nm *Manager) getProps(path dbus.ObjectPath, iface string) map[string]dbus.Variant {
	ac := nm.conn.Object("org.freedesktop.NetworkManager", path)

	var props map[string]dbus.Variant

	err := ac.Call("org.freedesktop.DBus.Properties.GetAll", 0, iface).Store(&props)
	if err != nil {
		panic(err)
	}

	return props
}

func (nm *Manager) getAccessPointProps(path dbus.ObjectPath) map[string]dbus.Variant {
	ap := nm.conn.Object("org.freedesktop.NetworkManager", path)

	var props map[string]dbus.Variant

	err := ap.Call("org.freedesktop.DBus.Properties.GetAll", 0, "org.freedesktop.NetworkManager.AccessPoint").Store(&props)
	if err != nil {
		panic(err)
	}

	return props
}

func (nm *Manager) getActiveConnections() []dbus.ObjectPath {
	obj := nm.conn.Object("org.freedesktop.NetworkManager", "/org/freedesktop/NetworkManager")

	var paths []dbus.ObjectPath

	err := obj.Call("org.freedesktop.DBus.Properties.Get", 0, "org.freedesktop.NetworkManager", "ActiveConnections").Store(&paths)
	if err != nil {
		panic(err)
	}

	return paths
}

func signalToPaths(signal *dbus.Signal) (paths []dbus.ObjectPath) {
	propmap := signal.Body[1].(map[string]dbus.Variant)
	paths = propmap["ActiveConnections"].Value().([]dbus.ObjectPath)
	return paths
}

func (nm *Manager) decodeActiveConnectionsSignal(signal *dbus.Signal) (conns []Connection, err error) {
	defer func() {
		if r := recover(); r != nil {
			conns = nil
			err = fmt.Errorf("recovered in decodeActiveConnectionsSignal: %v", r)
		}
	}()

	paths := signalToPaths(signal)
	return nm.pathsToConnections(paths), nil
}

type Connection struct {
	Type string
	Name string
	Rate uint32 // bits per second
	Good bool
}

func (nm *Manager) WatchActiveConnections() (<-chan []Connection, error) {
	nm.conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, "type='signal',sender='org.freedesktop.NetworkManager',path='/org/freedesktop/NetworkManager',interface='org.freedesktop.DBus.Properties',member='PropertiesChanged'")

	updates := make(chan []Connection)

	signals := make(chan *dbus.Signal)
	nm.conn.Signal(signals)

	go func() {
		for s := range signals {
			conns, err := nm.decodeActiveConnectionsSignal(s)
			if err != nil {
				continue
			}

			updates <- conns
		}
	}()

	return updates, nil
}

func (nm *Manager) pathsToConnections(paths []dbus.ObjectPath) []Connection {
	conns := make([]Connection, len(paths), len(paths))

	for i, path := range paths {
		props := nm.getProps(path, "org.freedesktop.NetworkManager.Connection.Active")
		conns[i].Type = props["Type"].Value().(string)
		conns[i].Name = props["Id"].Value().(string)
		conns[i].Good = props["State"].Value().(uint32) == 2

		if conns[i].Type == "802-11-wireless" {
			devPath := props["Devices"].Value().([]dbus.ObjectPath)[0]
			devProps := nm.getProps(devPath, "org.freedesktop.NetworkManager.Device.Wireless")
			accessPointPath := devProps["ActiveAccessPoint"].Value().(dbus.ObjectPath)
			accessPointProps := nm.getAccessPointProps(accessPointPath)
			// MaxBitrate is in kilobits per second
			conns[i].Rate = accessPointProps["MaxBitrate"].Value().(uint32) * 1000
		}
	}

	return conns
}

func (nm *Manager) GetActiveConnections() (conns []Connection, err error) {
	defer func() {
		if r := recover(); r != nil {
			conns = nil
			err = fmt.Errorf("recovered in GetActiveConnections: %v", r)
		}
	}()

	paths := nm.getActiveConnections()
	return nm.pathsToConnections(paths), nil
}

