let isRecording = false;
let recorder;
let audioStream;
const recordButton = document.getElementById('recordButton');

recordButton.addEventListener('click', async () => {
    if (isRecording) {
        // 停止录音
        recorder.stop();
        audioStream.getTracks().forEach(track => track.stop());
        
        // 导出音频并发送
        recorder.exportWAV(blob => {
            const formData = new FormData();
            formData.append('audio', blob, 'audio.wav');
            
            // 发送音频到服务器
            fetch('/upload', {
                method: 'POST',
                body: formData
            }).then(response => {
                console.log('音频已发送');
            }).catch(error => {
                console.error('发送音频时出错:', error);
            });
        });
        
        // 更新按钮状态
        isRecording = false;
        recordButton.classList.remove('recording'); // 恢复红色
        recordButton.innerText = '录音';
    } else {
        // 开始录音
        try {
            audioStream = await navigator.mediaDevices.getUserMedia({ audio: true });
            const audioContext = new (window.AudioContext || window.webkitAudioContext)();
            const input = audioContext.createMediaStreamSource(audioStream);
            recorder = new Recorder(input, { numChannels: 1 });
            recorder.record();
            
            // 更新按钮状态
            isRecording = true;
            recordButton.classList.add('recording'); // 改为绿色
            recordButton.innerText = '停止录音';
        } catch (error) {
            console.error('获取音频流失败:', error);
            alert('无法访问麦克风，权限问题或浏览器不支持');
        }
    }
});
