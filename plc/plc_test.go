// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package plc

import (
	"testing"

	"github.com/Team254/cheesy-arena/websocket"
	"github.com/goburrow/modbus"
	"github.com/stretchr/testify/assert"
)

func TestPlcInitialization(t *testing.T) {
	var client FakeModbusClient
	var plc ModbusPlc
	var notifier websocket.Notifier
	plc.client = &client
	plc.handler = modbus.NewTCPClientHandler("dummy")
	plc.ioChangeNotifier = &notifier

	assert.Equal(t, false, plc.IsEnabled())
	plc.SetAddress("dummy")
	assert.Equal(t, true, plc.IsEnabled())
	assert.Equal(t, &notifier, plc.IoChangeNotifier())
}

func TestPlcGetCycleState(t *testing.T) {
	var client FakeModbusClient
	var plc ModbusPlc
	plc.client = &client
	plc.handler = modbus.NewTCPClientHandler("dummy")
	plc.ioChangeNotifier = &websocket.Notifier{}

	assert.Equal(t, false, plc.GetCycleState(3, 1, 2))
	plc.update()
	assert.Equal(t, false, plc.GetCycleState(3, 1, 2))
	plc.update()
	assert.Equal(t, true, plc.GetCycleState(3, 1, 2))
	plc.update()
	assert.Equal(t, true, plc.GetCycleState(3, 1, 2))
	plc.update()
	assert.Equal(t, false, plc.GetCycleState(3, 1, 2))
	plc.update()
	assert.Equal(t, false, plc.GetCycleState(3, 1, 2))
	plc.update()

	assert.Equal(t, false, plc.GetCycleState(3, 1, 2))
	plc.update()
	assert.Equal(t, false, plc.GetCycleState(3, 1, 2))
	plc.update()
	assert.Equal(t, true, plc.GetCycleState(3, 1, 2))
	plc.update()
	assert.Equal(t, true, plc.GetCycleState(3, 1, 2))
	plc.update()
	assert.Equal(t, false, plc.GetCycleState(3, 1, 2))
	plc.update()
	assert.Equal(t, false, plc.GetCycleState(3, 1, 2))
}

func TestPlcGetNames(t *testing.T) {
	var plc ModbusPlc

	assert.Equal(
		t,
		[]string{
			"fieldEStop",
			"input(1)",
			"input(2)",
			"input(3)",
			"input(4)",
			"input(5)",
			"input(6)",
			"input(7)",
			"input(8)",
			"input(9)",
			"input(10)",
			"input(11)",
			"input(12)",
			"input(13)",
			"input(14)",
			"input(15)",
			"red1EStop",
			"red1AStop",
			"red2EStop",
			"red2AStop",
			"red3EStop",
			"red3AStop",
			"redConnected1",
			"redConnected2",
			"redConnected3",
			"input(25)",
			"input(26)",
			"input(27)",
			"input(28)",
			"input(29)",
			"input(30)",
			"input(31)",
			"blue1EStop",
			"blue1AStop",
			"blue2EStop",
			"blue2AStop",
			"blue3EStop",
			"blue3AStop",
			"blueConnected1",
			"blueConnected2",
			"blueConnected3",
		},
		plc.GetInputNames(),
	)

	// 2026: Updated to check for Fuel registers
	assert.Equal(
		t,
		[]string{
			"fieldIoConnection",
			"redFuel",
			"blueFuel",
		},
		plc.GetRegisterNames(),
	)

	// 2026: Updated to check for Hub Lights instead of Truss Lights
	assert.Equal(
		t,
		[]string{
			"heartbeat",
			"matchReset",
			"stackLightGreen",
			"stackLightOrange",
			"stackLightRed",
			"stackLightBlue",
			"stackLightBuzzer",
			"fieldResetLight",
			"redHubLight",
			"blueHubLight",
		},
		plc.GetCoilNames(),
	)
}

func TestPlcInputs(t *testing.T) {
	var client FakeModbusClient
	var plc ModbusPlc
	plc.client = &client
	plc.handler = modbus.NewTCPClientHandler("dummy")
	plc.ioChangeNotifier = &websocket.Notifier{}

	client.inputs[fieldEStop] = false
	plc.update()
	assert.Equal(t, false, plc.GetFieldEStop())
	client.inputs[fieldEStop] = true
	plc.update()
	assert.Equal(t, true, plc.GetFieldEStop())

	client.inputs[red1EStop] = false
	client.inputs[red1AStop] = false
	client.inputs[red2EStop] = false
	client.inputs[red2AStop] = false
	client.inputs[red3EStop] = false
	client.inputs[red3AStop] = false
	client.inputs[blue1EStop] = false
	client.inputs[blue1AStop] = false
	client.inputs[blue2EStop] = false
	client.inputs[blue2AStop] = false
	client.inputs[blue3EStop] = false
	client.inputs[blue3AStop] = false
	plc.update()
	redEStops, blueEStops := plc.GetTeamEStops()
	redAStops, blueAStops := plc.GetTeamAStops()
	assert.Equal(t, [3]bool{false, false, false}, redEStops)
	assert.Equal(t, [3]bool{false, false, false}, blueEStops)
	assert.Equal(t, [3]bool{false, false, false}, redAStops)
	assert.Equal(t, [3]bool{false, false, false}, blueAStops)
	client.inputs[red1EStop] = true
	plc.update()
	redEStops, blueEStops = plc.GetTeamEStops()
	redAStops, blueAStops = plc.GetTeamAStops()
	assert.Equal(t, [3]bool{true, false, false}, redEStops)
	assert.Equal(t, [3]bool{false, false, false}, blueEStops)
	assert.Equal(t, [3]bool{false, false, false}, redAStops)
	assert.Equal(t, [3]bool{false, false, false}, blueAStops)
	client.inputs[red1AStop] = true
	plc.update()
	redEStops, blueEStops = plc.GetTeamEStops()
	redAStops, blueAStops = plc.GetTeamAStops()
	assert.Equal(t, [3]bool{true, false, false}, redEStops)
	assert.Equal(t, [3]bool{false, false, false}, blueEStops)
	assert.Equal(t, [3]bool{true, false, false}, redAStops)
	assert.Equal(t, [3]bool{false, false, false}, blueAStops)
	client.inputs[red2EStop] = true
	plc.update()
	redEStops, blueEStops = plc.GetTeamEStops()
	redAStops, blueAStops = plc.GetTeamAStops()
	assert.Equal(t, [3]bool{true, true, false}, redEStops)
	assert.Equal(t, [3]bool{false, false, false}, blueEStops)
	assert.Equal(t, [3]bool{true, false, false}, redAStops)
	assert.Equal(t, [3]bool{false, false, false}, blueAStops)
	client.inputs[red2AStop] = true
	plc.update()
	redEStops, blueEStops = plc.GetTeamEStops()
	redAStops, blueAStops = plc.GetTeamAStops()
	assert.Equal(t, [3]bool{true, true, false}, redEStops)
	assert.Equal(t, [3]bool{false, false, false}, blueEStops)
	assert.Equal(t, [3]bool{true, true, false}, redAStops)
	assert.Equal(t, [3]bool{false, false, false}, blueAStops)
	client.inputs[red3EStop] = true
	plc.update()
	redEStops, blueEStops = plc.GetTeamEStops()
	redAStops, blueAStops = plc.GetTeamAStops()
	assert.Equal(t, [3]bool{true, true, true}, redEStops)
	assert.Equal(t, [3]bool{false, false, false}, blueEStops)
	assert.Equal(t, [3]bool{true, true, false}, redAStops)
	assert.Equal(t, [3]bool{false, false, false}, blueAStops)
	client.inputs[red3AStop] = true
	plc.update()
	redEStops, blueEStops = plc.GetTeamEStops()
	redAStops, blueAStops = plc.GetTeamAStops()
	assert.Equal(t, [3]bool{true, true, true}, redEStops)
	assert.Equal(t, [3]bool{false, false, false}, blueEStops)
	assert.Equal(t, [3]bool{true, true, true}, redAStops)
	assert.Equal(t, [3]bool{false, false, false}, blueAStops)
	client.inputs[blue1EStop] = true
	plc.update()
	redEStops, blueEStops = plc.GetTeamEStops()
	redAStops, blueAStops = plc.GetTeamAStops()
	assert.Equal(t, [3]bool{true, true, true}, redEStops)
	assert.Equal(t, [3]bool{true, false, false}, blueEStops)
	assert.Equal(t, [3]bool{true, true, true}, redAStops)
	assert.Equal(t, [3]bool{false, false, false}, blueAStops)
	client.inputs[blue1AStop] = true
	plc.update()
	redEStops, blueEStops = plc.GetTeamEStops()
	redAStops, blueAStops = plc.GetTeamAStops()
	assert.Equal(t, [3]bool{true, true, true}, redEStops)
	assert.Equal(t, [3]bool{true, false, false}, blueEStops)
	assert.Equal(t, [3]bool{true, true, true}, redAStops)
	assert.Equal(t, [3]bool{true, false, false}, blueAStops)
	client.inputs[blue2EStop] = true
	plc.update()
	redEStops, blueEStops = plc.GetTeamEStops()
	redAStops, blueAStops = plc.GetTeamAStops()
	assert.Equal(t, [3]bool{true, true, true}, redEStops)
	assert.Equal(t, [3]bool{true, true, false}, blueEStops)
	assert.Equal(t, [3]bool{true, true, true}, redAStops)
	assert.Equal(t, [3]bool{true, false, false}, blueAStops)
	client.inputs[blue2AStop] = true
	plc.update()
	redEStops, blueEStops = plc.GetTeamEStops()
	redAStops, blueAStops = plc.GetTeamAStops()
	assert.Equal(t, [3]bool{true, true, true}, redEStops)
	assert.Equal(t, [3]bool{true, true, false}, blueEStops)
	assert.Equal(t, [3]bool{true, true, true}, redAStops)
	assert.Equal(t, [3]bool{true, true, false}, blueAStops)
	client.inputs[blue3EStop] = true
	plc.update()
	redEStops, blueEStops = plc.GetTeamEStops()
	redAStops, blueAStops = plc.GetTeamAStops()
	assert.Equal(t, [3]bool{true, true, true}, redEStops)
	assert.Equal(t, [3]bool{true, true, true}, blueEStops)
	assert.Equal(t, [3]bool{true, true, true}, redAStops)
	assert.Equal(t, [3]bool{true, true, false}, blueAStops)
	client.inputs[blue3AStop] = true
	plc.update()
	redEStops, blueEStops = plc.GetTeamEStops()
	redAStops, blueAStops = plc.GetTeamAStops()
	assert.Equal(t, [3]bool{true, true, true}, redEStops)
	assert.Equal(t, [3]bool{true, true, true}, blueEStops)
	assert.Equal(t, [3]bool{true, true, true}, redAStops)
	assert.Equal(t, [3]bool{true, true, true}, blueAStops)

	client.inputs[redConnected1] = false
	client.inputs[redConnected2] = false
	client.inputs[redConnected3] = false
	client.inputs[blueConnected1] = false
	client.inputs[blueConnected2] = false
	client.inputs[blueConnected3] = false
	plc.update()
	redConnected, blueConnected := plc.GetEthernetConnected()
	assert.Equal(t, [3]bool{false, false, false}, redConnected)
	assert.Equal(t, [3]bool{false, false, false}, blueConnected)
	client.inputs[redConnected1] = true
	plc.update()
	redConnected, blueConnected = plc.GetEthernetConnected()
	assert.Equal(t, [3]bool{true, false, false}, redConnected)
	assert.Equal(t, [3]bool{false, false, false}, blueConnected)
	client.inputs[redConnected2] = true
	plc.update()
	redConnected, blueConnected = plc.GetEthernetConnected()
	assert.Equal(t, [3]bool{true, true, false}, redConnected)
	assert.Equal(t, [3]bool{false, false, false}, blueConnected)
	client.inputs[redConnected3] = true
	plc.update()
	redConnected, blueConnected = plc.GetEthernetConnected()
	assert.Equal(t, [3]bool{true, true, true}, redConnected)
	assert.Equal(t, [3]bool{false, false, false}, blueConnected)
	client.inputs[blueConnected1] = true
	plc.update()
	redConnected, blueConnected = plc.GetEthernetConnected()
	assert.Equal(t, [3]bool{true, true, true}, redConnected)
	assert.Equal(t, [3]bool{true, false, false}, blueConnected)
	client.inputs[blueConnected2] = true
	plc.update()
	redConnected, blueConnected = plc.GetEthernetConnected()
	assert.Equal(t, [3]bool{true, true, true}, redConnected)
	assert.Equal(t, [3]bool{true, true, false}, blueConnected)
	client.inputs[blueConnected3] = true
	plc.update()
	redConnected, blueConnected = plc.GetEthernetConnected()
	assert.Equal(t, [3]bool{true, true, true}, redConnected)
	assert.Equal(t, [3]bool{true, true, true}, blueConnected)
}

func TestPlcInputsGameSpecific(t *testing.T) {
	var client FakeModbusClient
	var plc ModbusPlc
	plc.client = &client
	plc.handler = modbus.NewTCPClientHandler("dummy")
	plc.ioChangeNotifier = &websocket.Notifier{}

	// None in 2026.
}

func TestPlcRegisters(t *testing.T) {
	var client FakeModbusClient
	var plc ModbusPlc
	plc.client = &client
	plc.handler = modbus.NewTCPClientHandler("dummy")
	plc.ioChangeNotifier = &websocket.Notifier{}

	testCases := map[uint16][4]bool{
		0:  {false, false, false, false},
		1:  {true, false, false, false},
		2:  {false, true, false, false},
		3:  {true, true, false, false},
		4:  {false, false, true, false},
		5:  {true, false, true, false},
		6:  {false, true, true, false},
		7:  {true, true, true, false},
		8:  {false, false, false, true},
		9:  {true, false, false, true},
		10: {false, true, false, true},
		11: {true, true, false, true},
		12: {false, false, true, true},
		13: {true, false, true, true},
		14: {false, true, true, true},
		15: {true, true, true, true},
	}

	for value, bits := range testCases {
		client.registers[0] = value
		plc.update()
		assert.Equal(
			t,
			map[string]bool{"RedDs": bits[0], "BlueDs": bits[1], "RedIoLink": bits[2], "BlueIoLink": bits[3]},
			plc.GetArmorBlockStatuses(),
		)
	}
}

// 2026: Updated test for Fuel Counts
func TestPlcRegistersGameSpecific(t *testing.T) {
	var client FakeModbusClient
	var plc ModbusPlc
	plc.client = &client
	plc.handler = modbus.NewTCPClientHandler("dummy")
	plc.ioChangeNotifier = &websocket.Notifier{}

	client.registers[1] = 0
	client.registers[2] = 0
	plc.update()
	redFuel, blueFuel := plc.GetFuelCounts()
	assert.Equal(t, 0, redFuel)
	assert.Equal(t, 0, blueFuel)

	client.registers[1] = 12 // redFuel register
	plc.update()
	redFuel, blueFuel = plc.GetFuelCounts()
	assert.Equal(t, 12, redFuel)
	assert.Equal(t, 0, blueFuel)

	client.registers[2] = 34 // blueFuel register
	plc.update()
	redFuel, blueFuel = plc.GetFuelCounts()
	assert.Equal(t, 12, redFuel)
	assert.Equal(t, 34, blueFuel)
}

func TestPlcCoils(t *testing.T) {
	var client FakeModbusClient
	var plc ModbusPlc
	plc.client = &client
	plc.handler = modbus.NewTCPClientHandler("dummy")
	plc.ioChangeNotifier = &websocket.Notifier{}

	assert.Equal(t, false, client.coils[0])
	plc.update()
	assert.Equal(t, true, client.coils[0])

	assert.Equal(t, false, client.coils[1])
	client.registers[fieldIoConnection] = 31
	plc.registers[fieldIoConnection] = 31
	plc.registers[redFuel] = 1
	plc.registers[blueFuel] = 2
	plc.ResetMatch()
	plc.update()
	assert.Equal(t, true, client.coils[1])
	assert.Equal(t, 31, int(plc.registers[fieldIoConnection]))
	assert.Equal(t, 0, int(plc.registers[redFuel]))
	assert.Equal(t, 0, int(plc.registers[blueFuel]))

	plc.SetStackLights(false, false, false, false)
	plc.update()
	assert.Equal(t, false, client.coils[2])
	assert.Equal(t, false, client.coils[3])
	assert.Equal(t, false, client.coils[4])
	assert.Equal(t, false, client.coils[5])
	plc.SetStackLights(true, false, false, false)
	plc.update()
	assert.Equal(t, false, client.coils[2])
	assert.Equal(t, false, client.coils[3])
	assert.Equal(t, true, client.coils[4])
	assert.Equal(t, false, client.coils[5])
	plc.SetStackLights(true, true, false, false)
	plc.update()
	assert.Equal(t, false, client.coils[2])
	assert.Equal(t, false, client.coils[3])
	assert.Equal(t, true, client.coils[4])
	assert.Equal(t, true, client.coils[5])
	plc.SetStackLights(true, true, true, false)
	plc.update()
	assert.Equal(t, false, client.coils[2])
	assert.Equal(t, true, client.coils[3])
	assert.Equal(t, true, client.coils[4])
	assert.Equal(t, true, client.coils[5])
	plc.SetStackLights(true, true, true, true)
	plc.update()
	assert.Equal(t, true, client.coils[2])
	assert.Equal(t, true, client.coils[3])
	assert.Equal(t, true, client.coils[4])
	assert.Equal(t, true, client.coils[5])

	plc.SetStackBuzzer(false)
	plc.update()
	assert.Equal(t, false, client.coils[6])
	plc.SetStackBuzzer(true)
	plc.update()
	assert.Equal(t, true, client.coils[6])

	plc.SetFieldResetLight(false)
	plc.update()
	assert.Equal(t, false, client.coils[7])
	plc.SetFieldResetLight(true)
	plc.update()
	assert.Equal(t, true, client.coils[7])
}

// 2026: Updated test for Hub Lights
func TestPlcCoilsGameSpecific(t *testing.T) {
	var client FakeModbusClient
	var plc ModbusPlc
	plc.client = &client
	plc.handler = modbus.NewTCPClientHandler("dummy")
	plc.ioChangeNotifier = &websocket.Notifier{}

	// Initial State: Off
	plc.SetHubLights(false, false)
	plc.update()
	assert.Equal(t, false, client.coils[8]) // redHubLight
	assert.Equal(t, false, client.coils[9]) // blueHubLight

	// Red On
	plc.SetHubLights(true, false)
	plc.update()
	assert.Equal(t, true, client.coils[8])
	assert.Equal(t, false, client.coils[9])

	// Blue On
	plc.SetHubLights(false, true)
	plc.update()
	assert.Equal(t, false, client.coils[8])
	assert.Equal(t, true, client.coils[9])

	// Both On
	plc.SetHubLights(true, true)
	plc.update()
	assert.Equal(t, true, client.coils[8])
	assert.Equal(t, true, client.coils[9])
}

func TestPlcIsHealthy(t *testing.T) {
	var client FakeModbusClient
	var plc ModbusPlc
	plc.client = &client
	plc.handler = modbus.NewTCPClientHandler("dummy")
	plc.ioChangeNotifier = &websocket.Notifier{}

	assert.Equal(t, false, plc.IsHealthy())
	plc.update()
	assert.Equal(t, true, plc.IsHealthy())

	client.returnError = true
	plc.update()
	assert.Equal(t, false, plc.IsHealthy())
	plc.update()
	assert.Equal(t, false, plc.IsHealthy())

	client.returnError = false
	plc.update()
	assert.Equal(t, false, plc.IsHealthy())
}

func TestByteToBool(t *testing.T) {
	bytes := []byte{7, 254, 3}
	bools := byteToBool(bytes, 17)
	if assert.Equal(t, 17, len(bools)) {
		expectedBools := []bool{
			true, true, true, false, false, false, false, false, false, true, true, true, true, true, true, true, true,
		}
		assert.Equal(t, expectedBools, bools)
	}
}

func TestByteToUint(t *testing.T) {
	bytes := []byte{1, 77, 2, 253, 21, 179}
	uints := byteToUint(bytes, 3)
	if assert.Equal(t, 3, len(uints)) {
		assert.Equal(t, []uint16{333, 765, 5555}, uints)
	}
}

func TestBoolToByte(t *testing.T) {
	bools := []bool{true, true, false, false, true, false, false, false, false, true}
	bytes := boolToByte(bools)
	if assert.Equal(t, 2, len(bytes)) {
		assert.Equal(t, []byte{19, 2}, bytes)
		assert.Equal(t, bools, byteToBool(bytes, len(bools)))
	}
}
