package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
     "math/rand"
     "path/filepath"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

type Config struct {
	Mqtt struct {
		Broker   string `json:"broker"`
		ClientID string `json:"client_id"`
		Username string `json:"username"`
		Password string `json:"password"`
		Topic    string `json:"topic"`
	} `json:"mqtt"`
	TTS struct {
		Voice      string `json:"voice"`
		Rate       string `json:"rate"`
		RequestURL string `json:"request_url"`
	} `json:"tts"`
	Web struct {
		Port            string `json:"port"`
		WebPath         string `json:"web_path"`
		UploadPath      string `json:"upload_path"`
		VolScalingFactor string `json:"vol_scaling_factor"`
	} `json:"web"`
	AudioAPI struct {
		PlayURL  string `json:"play_url"`
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"audio_api"`
}

var config Config

type WAVHeader struct {
	RIFFHeader    [4]byte
	RIFFSize      uint32
	WAVEHeader    [4]byte
	FMTHeader     [4]byte
	FMTSize       uint32
	AudioFormat   uint16
	NumChannels   uint16
	SampleRate    uint32
	ByteRate      uint32
	BlockAlign    uint16
	BitsPerSample uint16
	DataHeader    [4]byte
	DataSize      uint32
}

func loadConfig(configPath string) error {
	file, err := ioutil.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	if err := json.Unmarshal(file, &config); err != nil {
		return fmt.Errorf("解析配置文件失败: %v", err)
	}
	return nil
}

// MQTT处理部分
var messagePubHandler MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
	log.Printf("接收到MQTT消息: %s\n", msg.Payload())
	processText(string(msg.Payload()))
}

func processText(text string) {
    encodedText := url.QueryEscape(text)
    requestURL := fmt.Sprintf("%s?text=%s&voice=%s&rate=%s",
        config.TTS.RequestURL,
        encodedText,
        config.TTS.Voice,
        config.TTS.Rate)

    resp, err := http.Get(requestURL)
    if err != nil {
        log.Println("下载wav文件失败:", err)
        return
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        log.Println("下载失败，状态码:", resp.StatusCode)
        return
    }

    // 创建临时文件
    wavFile, err := ioutil.TempFile("/tmp/", "tts_*.wav")
    if err != nil {
        log.Println("创建临时文件失败:", err)
        return
    }
    defer func() {
        wavFile.Close()
        os.Remove(wavFile.Name())
    }()

    // 写入临时文件
    if _, err = io.Copy(wavFile, resp.Body); err != nil {
        log.Println("保存音频失败:", err)
        return
    }

    // 重新打开读取
    wavData, err := os.ReadFile(wavFile.Name())
    if err != nil {
        log.Println("读取音频失败:", err)
        return
    }

	req, err := http.NewRequest("POST", config.AudioAPI.PlayURL, bytes.NewReader(wavData))
	if err != nil {
		log.Println("创建上传请求失败:", err)
		return
	}

	req.Header.Set("Content-Type", "audio/wav")
	req.SetBasicAuth(config.AudioAPI.Username, config.AudioAPI.Password)

	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		log.Println("上传音频失败:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("上传音频失败，状态码:", resp.StatusCode)
		return
	}

	log.Println("TTS音频上传并播放成功")
}

// HTTP处理部分
func downsample(data []int16, oldRate, newRate int) []int16 {
	ratio := float64(oldRate) / float64(newRate)
	newSize := int(float64(len(data)) / ratio)
	downsampled := make([]int16, newSize)

	for i := 0; i < newSize; i++ {
		srcIdx := float64(i) * ratio
		leftIdx := int(srcIdx)
		rightIdx := leftIdx + 1

		if rightIdx >= len(data) {
			rightIdx = len(data) - 1
		}

		alpha := srcIdx - float64(leftIdx)
		downsampled[i] = int16((1-alpha)*float64(data[leftIdx]) + alpha*float64(data[rightIdx]))
	}

	return downsampled
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	file, _, err := r.FormFile("audio")
	if err != nil {
		log.Printf("读取上传文件失败: %v", err)
		http.Error(w, "无法读取文件", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Printf("读取文件字节失败: %v", err)
		http.Error(w, "无法读取文件", http.StatusInternalServerError)
		return
	}

	var header WAVHeader
	err = binary.Read(bytes.NewReader(fileBytes), binary.LittleEndian, &header)
	if err != nil {
		log.Printf("解析WAV头失败: %v", err)
		http.Error(w, "无效的WAV文件", http.StatusBadRequest)
		return
	}

	if string(header.RIFFHeader[:]) != "RIFF" || string(header.WAVEHeader[:]) != "WAVE" || string(header.DataHeader[:]) != "data" {
		http.Error(w, "无效的WAV文件头", http.StatusBadRequest)
		return
	}

	pcmDataStart := binary.Size(header)
	pcmData := make([]int16, header.DataSize/2)
	err = binary.Read(bytes.NewReader(fileBytes[pcmDataStart:]), binary.LittleEndian, &pcmData)
	if err != nil {
		log.Printf("提取PCM数据失败: %v", err)
		http.Error(w, "PCM数据提取错误", http.StatusBadRequest)
		return
	}

	downsampledData := downsample(pcmData, int(header.SampleRate), 8000)

	scalingFactor, err := strconv.ParseFloat(config.Web.VolScalingFactor, 64)
	if err != nil {
		log.Printf("解析音量系数失败: %v", err)
		http.Error(w, "服务器配置错误", http.StatusInternalServerError)
		return
	}

	for i := range downsampledData {
		downsampledData[i] = int16(float64(downsampledData[i]) * scalingFactor)
	}

	paddedData := make([]int16, 4000+len(downsampledData)+12000)
	copy(paddedData[4000:], downsampledData)

	paddedBytes := make([]byte, len(paddedData)*2)
	for i, sample := range paddedData {
		binary.LittleEndian.PutUint16(paddedBytes[i*2:i*2+2], uint16(sample))
	}

    // 生成唯一文件名
    timestamp := time.Now().Format("20060102-150405")
    outputFile := fmt.Sprintf("padded_%s_%d.pcm", timestamp, rand.Intn(1000))
    outputPath := filepath.Join(config.Web.UploadPath, outputFile)

    // 写入文件
    if err := os.WriteFile(outputPath, paddedBytes, 0644); err != nil {
        log.Printf("保存文件失败: %v", err)
        http.Error(w, "保存失败", http.StatusInternalServerError)
        return
    }
        // 添加两分钟自动清理（可选）
    go func() {
        time.Sleep(2 * time.Minute)
        os.Remove(outputPath)
    }()

	req, err := http.NewRequest("POST", config.AudioAPI.PlayURL, bytes.NewReader(paddedBytes))
	if err != nil {
		log.Printf("创建转发请求失败: %v", err)
		http.Error(w, "无法转发音频数据", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "audio/wav")
	req.SetBasicAuth(config.AudioAPI.Username, config.AudioAPI.Password)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("转发音频数据失败: %v", err)
		http.Error(w, "无法转发音频数据", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("接收端返回错误状态码: %d", resp.StatusCode)
		http.Error(w, "音频处理失败", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("音频处理并转发成功"))
}

func startMQTTClient() {
	opts := MQTT.NewClientOptions()
	opts.AddBroker(config.Mqtt.Broker)
	opts.SetClientID(config.Mqtt.ClientID)
	opts.SetUsername(config.Mqtt.Username)
	opts.SetPassword(config.Mqtt.Password)
	opts.SetDefaultPublishHandler(messagePubHandler)

	opts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe(config.Mqtt.Topic, 0, nil); token.Wait() && token.Error() != nil {
			log.Printf("订阅失败: %v", token.Error())
		}
	}

	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Printf("MQTT连接失败: %v", token.Error())
		return
	}
	log.Println("MQTT客户端已连接")
}

func startHTTPServer() {
	http.HandleFunc("/upload", uploadHandler)
	http.Handle("/", http.FileServer(http.Dir(config.Web.WebPath)))

	log.Printf("HTTP服务器启动在端口 %s", config.Web.Port)
	log.Fatal(http.ListenAndServe(":"+config.Web.Port, nil))
}
//网页文本输入
func generateHandler(w http.ResponseWriter, r *http.Request) {
    // 设置CORS头
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

 //   log.Printf("收到请求: %s %s", r.Method, r.URL.Path)

    if r.Method == http.MethodOptions {
  //      log.Println("处理OPTIONS预检请求")
        return
    }

    if r.Method != http.MethodPost {
        errMsg := "不支持的请求方法"
        log.Printf("%s: %s", errMsg, r.Method)
        http.Error(w, errMsg, http.StatusMethodNotAllowed)
        return
    }

    // 读取请求体
    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        errMsg := "读取请求体失败"
        log.Printf("%s: %v", errMsg, err)
        http.Error(w, errMsg, http.StatusBadRequest)
        return
    }
    defer r.Body.Close()

   // log.Printf("原始请求体: %s", string(body))

    // 解析JSON
    var request struct {
        Text string `json:"text"`
    }
    if err := json.Unmarshal(body, &request); err != nil {
        errMsg := "JSON解析失败"
    //    log.Printf("%s: %v\n原始内容: %s", errMsg, err, string(body))
        http.Error(w, errMsg, http.StatusBadRequest)
        return
    }

    log.Printf("接收到网页文本内容: %q", request.Text)

    // 处理文本
    go func(text string) {
     //   log.Printf("开始异步处理文本，长度: %d 字符", len(text))
        processText(text)
    //    log.Printf("文本处理完成: %q", text)
    }(request.Text)


}

func main() {
	configPath := flag.String("c", "config.json", "配置文件路径")
	debug := flag.Bool("debug", false, "启用调试模式")
	textPtr := flag.String("p", "", "要转换为语音的文本")
	flag.Parse()

	if *debug {
		log.Println("调试模式已激活")
	}

	if err := loadConfig(*configPath); err != nil {
		log.Fatal(err)
	}

	if *textPtr != "" {
		processText(*textPtr)
	}
     http.HandleFunc("/generate", generateHandler)
	go startMQTTClient()
	startHTTPServer()
}