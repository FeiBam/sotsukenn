package com.example.myapp.model

import com.google.gson.annotations.SerializedName

/**
 * FCM Token 提交请求
 */
data class FcmTokenRequest(
    @SerializedName("token")
    val token: String,

    @SerializedName("device_name")
    val deviceName: String
)

/**
 * FCM Token 提交响应
 */
data class FcmTokenResponse(
    @SerializedName("success")
    val success: Boolean,

    @SerializedName("message")
    val message: String? = null
)
