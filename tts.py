#!/usr/bin/env python3

from flask import Flask, send_file, request, Response
import edge_tts
import io
import re


app = Flask(__name__)


@app.route("/api/ra", methods=["GET"])
def generate_audio():
    # 从查询参数中获取 text 和 voice
    text = request.args.get("text", default="说点什么吧", type=str)
    voice = request.args.get("voice", default="zh-CN-XiaoxiaoNeural", type=str)
    rate = request.args.get("rate", default="0")  # 默认值为 0%

    centerRate = 25  # 阅读的语速中间值
    try:
        rate_value = int(rate)  # 尝试将 rate 转换为整数
    except ValueError:
        rate_value = 0  # 如果转换失败，则将 rate 设置为 0

    if rate_value < centerRate:
        custom_rate = "-" + str((centerRate - rate_value)) + "%"  # 低于 50%
    else:
        custom_rate = "+" + str((rate_value - centerRate)) + "%"  # 高于 50%
    # 使用 edge_tts 生成音频
    communicate = edge_tts.Communicate(text, voice, rate=custom_rate)

    audio_stream = io.BytesIO()

    for chunk in communicate.stream_sync():
        if chunk["type"] == "audio":
            audio_stream.write(chunk["data"])

    audio_stream.seek(0)  # Reset the stream to the beginning

    return send_file(
        audio_stream,
        mimetype="audio/mpeg",
        as_attachment=True,
        download_name="audio.mp3",
    )   


@app.route("/api/rap", methods=["POST"])
def generate_audiop():
    # 从表单数据中获取 text 和 voice
    text = request.form.get("text", default="")
    voice = request.form.get("voice", default="zh-CN-XiaoxiaoNeural")
    rate = request.form.get("rate", default="0")  # 默认值为 0%
    centerRate = 25  # 阅读的语速中间值
    try:
        rate_value = int(rate)  # 尝试将 rate 转换为整数
    except ValueError:
        rate_value = 0  # 如果转换失败，则将 rate 设置为 0

    if rate_value < centerRate:
        custom_rate = "-" + str((centerRate - rate_value)) + "%"  # 低于 50%
    else:
        custom_rate = "+" + str((rate_value - centerRate)) + "%"  # 高于 50%

    # 使用 edge_tts 生成音频
    communicate = edge_tts.Communicate(text, voice, rate=custom_rate)
    audio_stream = io.BytesIO()

    for chunk in communicate.stream_sync():
        if chunk["type"] == "audio":
            audio_stream.write(chunk["data"])

    audio_stream.seek(0)  # Reset the stream to the beginning

    return send_file(
        audio_stream,
        mimetype="audio/mpeg",
        as_attachment=True,
        download_name="audio.mp3",
    )




def split_chat(text):
    # 使用正则表达式匹配对话部分
    dialogue_pattern = dialogue_pattern = r'“([^”]*[。！？，；])”'
    dialogues = re.finditer(dialogue_pattern, text)

    if not dialogues:
        return [[dialogues, 0]]

    # 初始化结果列表
    result = []

    # 上一个对话结束的位置
    last_end = 0

    # 遍历每个对话
    for match in dialogues:
        # 获取对话前的旁白
        pre_dialogue_text = text[last_end:match.start()].strip()
        if pre_dialogue_text:
            result.append([pre_dialogue_text, 0])
        
        # 获取对话内容
        dialogue_text = match.group(0).strip()
        if dialogue_text:
            result.append([dialogue_text, 1])
        
        # 更新最后一个对话结束的位置
        last_end = match.end()

    # 获取最后一个对话后的旁白
    post_dialogue_text = text[last_end:].strip()
    if post_dialogue_text:
        result.append([post_dialogue_text, 0])

    return result


@app.route("/api/rap2", methods=["POST"])
def generate_audiop2():
    # 从表单数据中获取 text 和 voice
    text = request.form.get("text", default="")
    voice = request.form.get("voice", default="zh-CN-XiaoxiaoNeural")  # 旁白
    voice_chat = request.form.get("voice_chat", default="zh-CN-YunxiNeural")  # 对话
    volume_chat = request.form.get("volume_chat", default="0")  # 对话的音量调整
    rate = request.form.get("rate", default="0")  # 默认值为 0%
    centerRate = 25  # 阅读的语速中间值
    try:
        rate_value = int(rate)  # 尝试将 rate 转换为整数
    except ValueError:
        rate_value = 0  # 如果转换失败，则将 rate 设置为 0

    if rate_value < centerRate:
        custom_rate = "-" + str((centerRate - rate_value)) + "%"  # 低于 50%
    else:
        custom_rate = "+" + str((rate_value - centerRate)) + "%"  # 高于 50%
    
    try:
        volume_chat = int(volume_chat)  # 尝试将 volume_chat 转换为整数
    except ValueError:
        volume_chat = 0  # 如果转换失败，则将 volume_chat 设置为 0
    volume_chat = "+" + str(volume_chat) + "%"

    # 使用 edge_tts 生成音频
    audio_stream = io.BytesIO()

    text_splitted = split_chat(text)
    for text_seg in text_splitted:
        if text_seg[1] == 0:
            communicate = edge_tts.Communicate(text_seg[0], voice, rate=custom_rate)
        else:
            communicate = edge_tts.Communicate(text_seg[0], voice_chat, rate=custom_rate, volume=volume_chat)
        for chunk in communicate.stream_sync():
            if chunk["type"] == "audio":
                audio_stream.write(chunk["data"])

    audio_stream.seek(0)  # Reset the stream to the beginning

    return send_file(
        audio_stream,
        mimetype="audio/mpeg",
        as_attachment=True,
        download_name="audio.mp3",
    )
import subprocess

@app.route("/api/ra_wav", methods=["GET"])
def generate_audio_wav():
    # 从查询参数中获取 text、voice 和 rate
    text = request.args.get("text", default="说点什么吧", type=str)
    voice = request.args.get("voice", default="zh-CN-XiaoxiaoNeural", type=str)
    rate = request.args.get("rate", default="0")  # 默认值为 0%

    centerRate = 25  # 语速中间值
    try:
        rate_value = int(rate)
    except ValueError:
        rate_value = 0

    if rate_value < centerRate:
        custom_rate = "-" + str((centerRate - rate_value)) + "%"
    else:
        custom_rate = "+" + str((rate_value - centerRate)) + "%"

    # 使用 edge_tts 生成 mp3 格式的音频流
    communicate = edge_tts.Communicate(text, voice, rate=custom_rate)
    audio_stream = io.BytesIO()
    for chunk in communicate.stream_sync():
        if chunk["type"] == "audio":
            audio_stream.write(chunk["data"])
    audio_stream.seek(0)

    # 使用 sox 将 mp3 数据转换为采样率为 8000Hz 的 wav 格式
    # sox 命令解释：
    # - "-t mp3": 指定输入格式为 mp3，通过标准输入读取
    # - "-": 表示从标准输入读取
    # - "-r 8000": 设置输出采样率为 8000Hz
    # - "-t wav": 指定输出格式为 wav，通过标准输出输出
    # - "-": 表示输出到标准输出
    process = subprocess.Popen(
        ["sox", "-t", "mp3", "-", "-r", "8000", "-t", "wav", "-"],
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE
    )
    output, error = process.communicate(audio_stream.read())
    if process.returncode != 0:
        return Response("音频转换失败: " + error.decode("utf-8"), status=500)

    wav_stream = io.BytesIO(output)
    wav_stream.seek(0)

    return send_file(
        wav_stream,
        mimetype="audio/wav",
        as_attachment=True,
        download_name="audio.wav",
    )

      #http://192.168.100.18:1233/api/ra_wav?text=%E4%BD%A0%E5%A5%BD&voice=zh-CN-YunxiNeural&rate=52

if __name__ == "__main__":
    app.run(host="0.0.0.0", port=1233)