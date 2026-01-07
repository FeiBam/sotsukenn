package com.example.myapp.model

import com.google.gson.annotations.SerializedName

/**
 * 摄像头列表响应
 */
data class CameraListResponse(
    @SerializedName("status")
    val status: String,
    @SerializedName("code")
    val code: Int,
    @SerializedName("error")
    val error: String,
    @SerializedName("message")
    val message: String,
    @SerializedName("body")
    val body: CameraListBody
)

data class CameraListBody(
    @SerializedName("cameras")
    val cameras: List<String>
)

/**
 * 摄像头快照响应
 */
data class CameraSnapshotResponse(
    @SerializedName("status")
    val status: String,
    @SerializedName("code")
    val code: Int,
    @SerializedName("error")
    val error: String,
    @SerializedName("message")
    val message: String,
    @SerializedName("body")
    val body: CameraSnapshotBody
)

data class CameraSnapshotBody(
    @SerializedName("camera_name")
    val cameraName: String,
    @SerializedName("url")
    val url: String,
    @SerializedName("frigate_token")
    val frigateToken: String
)

/**
 * 摄像头流 URL 响应
 */
data class CameraStreamResponse(
    @SerializedName("status")
    val status: String,
    @SerializedName("code")
    val code: Int,
    @SerializedName("error")
    val error: String,
    @SerializedName("message")
    val message: String,
    @SerializedName("body")
    val body: CameraStreamBody
)

data class CameraStreamBody(
    @SerializedName("camera_name")
    val cameraName: String,
    @SerializedName("stream_type")
    val streamType: String,
    @SerializedName("url")
    val url: String,
    @SerializedName("frigate_token")
    val frigateToken: String
)
