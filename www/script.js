let isRecording = false;
let recorder;
let audioStream;
const recordButton = document.getElementById('recordButton');

// 统一通知函数（需保持与HTML中定义一致）
function showNotification(message, type = 'success') {
    const notification = document.createElement('div');
    notification.className = `notification ${type}`;
    
    notification.innerHTML = `
        <div class="notification-icon"></div>
        <span>${message}</span>
        <div class="notification-progress"></div>
    `;

    document.body.appendChild(notification);
    
    // 自动移除逻辑
    setTimeout(() => {
        notification.style.animation = 'slideUp 0.3s forwards';
        setTimeout(() => notification.remove(), 300);
    }, 2700);
}

recordButton.addEventListener('click', async () => {
    if (isRecording) {
        // 停止录音
        recorder.stop();
        audioStream.getTracks().forEach(track => track.stop());
        
        // 显示处理中通知
        showNotification('音频处理中...', 'processing');

        // 导出音频并发送
        recorder.exportWAV(blob => {
            const formData = new FormData();
            formData.append('audio', blob, 'audio.wav');
            
            // 发送音频到服务器
            fetch('/upload', {
                method: 'POST',
                body: formData
            }).then(response => {
                if (!response.ok) throw new Error('上传失败');
                showNotification('音频上传成功', 'success');
            }).catch(error => {
                console.error('发送音频时出错:', error);
                showNotification(`上传失败: ${error.message}`, 'error');
            });
        });
        
        // 更新按钮状态
        isRecording = false;
        recordButton.classList.remove('recording');
        recordButton.innerText = '录音';
    } else {
        // 开始录音
        try {
            showNotification('准备录音设备...', 'processing');
            
            audioStream = await navigator.mediaDevices.getUserMedia({ audio: true });
            const audioContext = new (window.AudioContext || window.webkitAudioContext)();
            const input = audioContext.createMediaStreamSource(audioStream);
            recorder = new Recorder(input, { numChannels: 1 });
            recorder.record();
            
            // 更新按钮状态
            isRecording = true;
            recordButton.classList.add('recording');
            recordButton.innerText = '停止录音';
            
            showNotification('录音进行中...', 'processing');

        } catch (error) {
            console.error('获取音频流失败:', error);
            showNotification(`麦克风访问失败: ${error.message}`, 'error');
        }
    }
});