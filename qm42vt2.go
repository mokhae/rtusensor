package rtusensor

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/dpapathanasiou/go-modbus"
	"io/ioutil"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

var (
	task_time     int
	addr          string
	sensor_id     byte
	BaudRate      int
	DataBits      int
	StopBits      int
	logDir        string
	logginglevel  string
	outFile       string
	outFlag       bool
	sensorTimeout float64
	startAddress  uint16
	quantity      uint16
)

type qm42vt2 struct {
	Z_rms_velocity_in_per_sec    string `json:"z_rms_velocity_in_per_sec"`
	Z_rms_velocity_mm_per_sec    string `json:"z_rms_velocity_mm_per_sec"`
	Temp_dF                      string `json:"temp_dF"`
	Temp_dC                      string `json:"temp_dC"`
	X_rms_velocity_in_per_sec    string `json:"x_rms_velocity_in_per_sec"`
	X_rms_velocity_mm_per_sec    string `json:"x_rms_velocity_mm_per_sec"`
	Z_peak_acceleration          string `json:"z_peak_acceleration"`
	X_peak_acceleration          string `json:"x_peak_acceleration"`
	Z_peak_velocity_comp_freq    string `json:"z_peak_velocity_comp_freq"`
	X_peak_velocity_comp_freq    string `json:"x_peak_velocity_comp_freq"`
	Z_rms_acceleration           string `json:"z_rms_acceleration"`
	X_rms_acceleration           string `json:"x_rms_acceleration"`
	Z_kurtosis                   string `json:"z_kurtosis"`
	X_kurtosis                   string `json:"x_kurtosis"`
	Z_crest                      string `json:"z_crest"`
	X_crest                      string `json:"x_crest"`
	Z_peak_velocity_in_per_sec   string `json:"z_peak_velocity_in_per_sec"`
	Z_peak_velocity_mm_per_sec   string `json:"z_peak_velocity_mm_per_sec"`
	X_peak_velocity_in_per_sec   string `json:"x_peak_velocity_in_per_sec"`
	X_peak_velocity_mm_per_sec   string `json:"x_peak_velocity_mm_per_sec"`
	Z_high_freq_rms_acceleration string `json:"z_high_freq_rms_acceleration"`
	X_high_freq_rms_acceleration string `json:"x_high_freq_rms_acceleration"`
}

type Vibration struct {
	Id    byte    `json:"id"`
	Stime string  `json:"stime"`
	Data  qm42vt2 `json:"data"`
}
type MBClient struct {
	Name          string
	Configdata    *map[string]interface{}
	Buffer        chan string
	useTurckCloud bool
}

func convStringData(data uint16, decimalNum uint16, decimalPoint uint16) string {

	var ret string
	floatData := float64(data) / float64(decimalNum)
	//log.Println("Data : ", floatData)
	switch decimalPoint {
	case 1:
		ret = fmt.Sprintf("%.1f", floatData)
	case 2:
		ret = fmt.Sprintf("%.2f", floatData)
	case 3:
		ret = fmt.Sprintf("%.3f", floatData)
	case 4:
		ret = fmt.Sprintf("%.4f", floatData)
	default:
		ret = fmt.Sprintf("%.f", floatData)
	}

	return ret

}

func New(name string, configdata *map[string]interface{}, buffer *chan string, useTurckCloudFlag *bool) *MBClient {
	return &MBClient{Name: name, Configdata: configdata, Buffer: *buffer, useTurckCloud: *useTurckCloudFlag}
}

func (m MBClient) Run() {

	config := *m.Configdata

	task_time, _ = strconv.Atoi(fmt.Sprintf("%v", config["task_update_interval_ms"]))
	addr = fmt.Sprintf("%v", config["modbus_port"])

	//sensorId32, _ := strconv.ParseUint(fmt.Sprintf("%v",config["sensor_id"]), 10, 32)
	//sensor_id = byte(sensorId32)
	BaudRate, _ = strconv.Atoi(fmt.Sprintf("%v", config["sensor_baudrate"]))
	DataBits, _ = strconv.Atoi(fmt.Sprintf("%v", config["sensor_bytesize"]))
	StopBits, _ = strconv.Atoi(fmt.Sprintf("%v", config["sensor_stopbits"]))
	logDir = fmt.Sprintf("%v", config["logging_dir"])
	logginglevel = fmt.Sprintf("%v", config["logging_level"])
	outFile = fmt.Sprintf("%v", config["output_file_json"])
	outFlag, _ = strconv.ParseBool(fmt.Sprintf("%v", config["write_output_file_json"]))
	sensorTimeout, _ = strconv.ParseFloat(fmt.Sprintf("%v", config["sensor_timeout"]), 64)

	startAddress32, _ := strconv.ParseUint(fmt.Sprintf("%v", config["start_address"]), 10, 32)
	startAddress = uint16(startAddress32)
	quantity32, _ := strconv.ParseUint(fmt.Sprintf("%v", config["quantity"]), 10, 32)

	quantity = uint16(quantity32)

	log.Println("Name ", m.Name)
	log.Println("Address ", addr)
	//log.Println("ID ", sensor_id)
	log.Println("BaudRate ", BaudRate)
	log.Println("task_time ", task_time)
	log.Println("logging_dir ", logDir)
	log.Println("output_file_json ", outFile)
	log.Println("write_output_file_json ", outFlag)
	log.Println("sensor_timeout ", sensorTimeout)
	log.Println("start_address ", startAddress)
	log.Println("quantity ", quantity)

	//handler := modbus.NewRTUClientHandler(addr)
	//handler.BaudRate = 115200
	//handler.DataBits = 8
	//handler.Parity = "N"
	//handler.StopBits = 1
	////handler.SlaveId = 1
	//handler.Timeout = 2 * time.Second

	//timeOut := time.Duration(sensorTimeout) * time.Millisecond
	ctx, cerr := modbusclient.ConnectRTU(addr, BaudRate)
	if cerr != nil {
		log.Println(fmt.Sprintf("RTU Connection error: %s", cerr))
	} else {
		ticker := time.NewTicker(time.Millisecond * time.Duration(task_time))
		for {
			var sid string = fmt.Sprintf("%v", config["sensor_id"])

			sensorid := strings.Split(sid, ",")
			if len(sensorid) > 0 {
				for i := 0; i < len(sensorid); i++ {
					sensor_id, _ := strconv.ParseUint(strings.TrimSpace(sensorid[i]), 10, 32)
					id := byte(sensor_id)
					log.Println("ID ", id)
					/*
						handler.SlaveId = id

						client := modbus.NewClient(handler)
						log.Println("Modbus RTU Connected ")
						time.Sleep(time.Millisecond * 200)
						results, rerr := client.ReadHoldingRegisters(startAddress, quantity)
						if rerr != nil || results == nil {
							log.Println(fmt.Sprintf("The First Reading error: %s", rerr))
							//wg.Done()
							return
						}
						log.Println(fmt.Sprintf("Modbus RTU Reading : %d", len(results)))


					*/
					trace := true
					var responsePause int = 100

					select {

					case <-ticker.C:
						log.Println(fmt.Sprintf("Reading Start: %d", id))
						results, readErr := modbusclient.RTURead(ctx, id, modbusclient.FUNCTION_READ_HOLDING_REGISTERS, startAddress, quantity, responsePause, trace)
						if readErr != nil {
							log.Println(fmt.Sprintf("Reading error: %s", readErr))
							continue
						}

						// Data
						//log.Println("Len : " , len(results))
						//log.Println("data : " , results)
						var i int = 3
						qm42 := qm42vt2{
							Z_rms_velocity_in_per_sec:    convStringData(binary.BigEndian.Uint16([]byte{results[i+0], results[i+1]}), 10000, 4),
							Z_rms_velocity_mm_per_sec:    convStringData(binary.BigEndian.Uint16([]byte{results[i+2], results[i+3]}), 1000, 3),
							Temp_dF:                      convStringData(binary.BigEndian.Uint16([]byte{results[i+4], results[i+5]}), 100, 2),
							Temp_dC:                      convStringData(binary.BigEndian.Uint16([]byte{results[i+6], results[i+7]}), 100, 2),
							X_rms_velocity_in_per_sec:    convStringData(binary.BigEndian.Uint16([]byte{results[i+8], results[i+9]}), 10000, 4),
							X_rms_velocity_mm_per_sec:    convStringData(binary.BigEndian.Uint16([]byte{results[i+10], results[i+11]}), 1000, 3),
							Z_peak_acceleration:          convStringData(binary.BigEndian.Uint16([]byte{results[i+12], results[i+13]}), 1000, 3),
							X_peak_acceleration:          convStringData(binary.BigEndian.Uint16([]byte{results[i+14], results[i+15]}), 1000, 3),
							Z_peak_velocity_comp_freq:    convStringData(binary.BigEndian.Uint16([]byte{results[i+16], results[i+17]}), 10, 1),
							X_peak_velocity_comp_freq:    convStringData(binary.BigEndian.Uint16([]byte{results[i+18], results[i+19]}), 10, 1),
							Z_rms_acceleration:           convStringData(binary.BigEndian.Uint16([]byte{results[i+20], results[i+21]}), 1000, 3),
							X_rms_acceleration:           convStringData(binary.BigEndian.Uint16([]byte{results[i+22], results[i+23]}), 1000, 3),
							Z_kurtosis:                   convStringData(binary.BigEndian.Uint16([]byte{results[i+24], results[i+25]}), 1000, 3),
							X_kurtosis:                   convStringData(binary.BigEndian.Uint16([]byte{results[i+26], results[i+27]}), 1000, 3),
							Z_crest:                      convStringData(binary.BigEndian.Uint16([]byte{results[i+28], results[i+29]}), 1000, 3),
							X_crest:                      convStringData(binary.BigEndian.Uint16([]byte{results[i+30], results[i+31]}), 1000, 3),
							Z_peak_velocity_in_per_sec:   convStringData(binary.BigEndian.Uint16([]byte{results[i+32], results[i+33]}), 10000, 4),
							Z_peak_velocity_mm_per_sec:   convStringData(binary.BigEndian.Uint16([]byte{results[i+34], results[i+35]}), 1000, 3),
							X_peak_velocity_in_per_sec:   convStringData(binary.BigEndian.Uint16([]byte{results[i+36], results[i+37]}), 10000, 4),
							X_peak_velocity_mm_per_sec:   convStringData(binary.BigEndian.Uint16([]byte{results[i+38], results[i+39]}), 1000, 3),
							Z_high_freq_rms_acceleration: convStringData(binary.BigEndian.Uint16([]byte{results[i+40], results[i+41]}), 1000, 3),
							X_high_freq_rms_acceleration: convStringData(binary.BigEndian.Uint16([]byte{results[i+42], results[i+43]}), 1000, 3),
						}
						//log.Println("Z Axis Velocity : ", qm42.x_high_freq_rms_acceleration)
						var file []byte
						var err error
						if m.useTurckCloud {
							file, err = json.MarshalIndent(qm42, "", " ")
							if err != nil {
								log.Fatal(err)
							}
						} else {
							var vibData *Vibration = &Vibration{
								Id:    id,
								Stime: time.Now().Format("2006-01-01 15:04:05"),
								Data:  qm42,
							}
							file, err = json.MarshalIndent(vibData, "", " ")
							if err != nil {
								log.Fatal(err)
							}
						}

						//log.Println("Modbus ID : ", id, string(file))

						m.Buffer <- string(file)

						if outFlag {

							_ = ioutil.WriteFile(outFile, file, 0644)
						}

					}

					//time.Sleep(time.Millisecond * 1000)
				}
			}
		}

	}

}

func (m MBClient) RunSimulation() {
	config := *m.Configdata

	task_time, _ = strconv.Atoi(fmt.Sprintf("%v", config["task_update_interval_ms"]))
	addr = fmt.Sprintf("%v", config["modbus_port"])

	//sensorId32, _ := strconv.ParseUint(fmt.Sprintf("%v",config["sensor_id"]), 10, 32)
	//sensor_id = byte(sensorId32)
	BaudRate, _ = strconv.Atoi(fmt.Sprintf("%v", config["sensor_baudrate"]))
	DataBits, _ = strconv.Atoi(fmt.Sprintf("%v", config["sensor_bytesize"]))
	StopBits, _ = strconv.Atoi(fmt.Sprintf("%v", config["sensor_stopbits"]))
	logDir = fmt.Sprintf("%v", config["logging_dir"])
	logginglevel = fmt.Sprintf("%v", config["logging_level"])
	outFile = fmt.Sprintf("%v", config["output_file_json"])
	outFlag, _ = strconv.ParseBool(fmt.Sprintf("%v", config["write_output_file_json"]))
	sensorTimeout, _ = strconv.ParseFloat(fmt.Sprintf("%v", config["sensor_timeout"]), 64)

	startAddress32, _ := strconv.ParseUint(fmt.Sprintf("%v", config["start_address"]), 10, 32)
	startAddress = uint16(startAddress32)
	quantity32, _ := strconv.ParseUint(fmt.Sprintf("%v", config["quantity"]), 10, 32)
	quantity = uint16(quantity32)

	log.Println("Name ", m.Name)
	log.Println("Address ", addr)
	//log.Println("ID ", sensor_id)
	log.Println("BaudRate ", BaudRate)
	log.Println("task_time ", task_time)
	log.Println("logging_dir ", logDir)
	log.Println("output_file_json ", outFile)
	log.Println("write_output_file_json ", outFlag)
	log.Println("sensor_timeout ", sensorTimeout)
	log.Println("start_address ", startAddress)
	log.Println("quantity ", quantity)

	//handler := modbus.NewRTUClientHandler(addr)
	//handler.BaudRate = 115200
	//handler.DataBits = 8
	//handler.Parity = "N"
	//handler.StopBits = 1
	////handler.SlaveId = 1
	//handler.Timeout = 2 * time.Second

	ticker := time.NewTicker(time.Millisecond * time.Duration(task_time))
	for {
		var sid string = fmt.Sprintf("%v", config["sensor_id"])

		sensorid := strings.Split(sid, ",")
		if len(sensorid) > 0 {
			for i := 0; i < len(sensorid); i++ {
				sensor_id, _ := strconv.ParseUint(strings.TrimSpace(sensorid[i]), 10, 32)
				id := byte(sensor_id)
				log.Println("ID ", id)

				select {

				case <-ticker.C:
					log.Println(fmt.Sprintf("Reading Start: %d", id))

					results := randUint16(0, 65535, 22)

					qm42 := qm42vt2{
						Z_rms_velocity_in_per_sec:    convStringData(results[0], 10000, 4),
						Z_rms_velocity_mm_per_sec:    convStringData(results[1], 1000, 3),
						Temp_dF:                      convStringData(results[2], 100, 2),
						Temp_dC:                      convStringData(results[3], 100, 2),
						X_rms_velocity_in_per_sec:    convStringData(results[4], 10000, 4),
						X_rms_velocity_mm_per_sec:    convStringData(results[5], 1000, 3),
						Z_peak_acceleration:          convStringData(results[6], 1000, 3),
						X_peak_acceleration:          convStringData(results[7], 1000, 3),
						Z_peak_velocity_comp_freq:    convStringData(results[8], 10, 1),
						X_peak_velocity_comp_freq:    convStringData(results[9], 10, 1),
						Z_rms_acceleration:           convStringData(results[10], 1000, 3),
						X_rms_acceleration:           convStringData(results[11], 1000, 3),
						Z_kurtosis:                   convStringData(results[12], 1000, 3),
						X_kurtosis:                   convStringData(results[13], 1000, 3),
						Z_crest:                      convStringData(results[14], 1000, 3),
						X_crest:                      convStringData(results[15], 1000, 3),
						Z_peak_velocity_in_per_sec:   convStringData(results[16], 10000, 4),
						Z_peak_velocity_mm_per_sec:   convStringData(results[17], 1000, 3),
						X_peak_velocity_in_per_sec:   convStringData(results[18], 10000, 4),
						X_peak_velocity_mm_per_sec:   convStringData(results[19], 1000, 3),
						Z_high_freq_rms_acceleration: convStringData(results[20], 1000, 3),
						X_high_freq_rms_acceleration: convStringData(results[21], 1000, 3),
					}
					//log.Println("Z Axis Velocity : ", qm42.x_high_freq_rms_acceleration)
					var vibData *Vibration = &Vibration{
						Id:    id,
						Stime: time.Now().Format("2006-01-01 15:04:05.999"),
						Data:  qm42,
					}
					file, err := json.MarshalIndent(vibData, "", " ")
					if err != nil {
						log.Fatal(err)
					}
					//log.Println("Modbus ID : ", id, string(file))

					m.Buffer <- string(file)
					if outFlag {

						_ = ioutil.WriteFile(outFile, file, 0644)
					}

				}

				//time.Sleep(time.Millisecond * 1000)
			}
		}
	}
}

func randFloats(min, max float64, n int) []float64 {
	res := make([]float64, n)
	for i := range res {
		res[i] = min + rand.Float64()*(max-min)
	}
	return res
}

func randUint16(min, max int, n int) []uint16 {
	res := make([]uint16, n)
	for i := range res {
		res[i] = uint16(rand.Intn(max-min+1)) + uint16(min)
	}
	return res
}
