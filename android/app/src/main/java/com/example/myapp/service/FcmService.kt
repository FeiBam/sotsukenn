package com.example.myapp.service

import android.util.Log
import com.example.myapp.model.FcmTokenRequest
import com.example.myapp.network.ApiClient
import com.example.myapp.network.AuthService
import com.example.myapp.utils.DeviceInfoUtils
import com.example.myapp.utils.PreferenceManager
import com.example.myapp.utils.NotificationHelper
import com.google.firebase.messaging.FirebaseMessagingService
import com.google.firebase.messaging.RemoteMessage
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch

class FcmService : FirebaseMessagingService() {

    private lateinit var preferenceManager: PreferenceManager
    private lateinit var notificationHelper: NotificationHelper
    private val serviceScope = CoroutineScope(Dispatchers.IO)

    companion object {
        private const val TAG = "FcmService"
    }

    override fun onCreate() {
        super.onCreate()
        preferenceManager = PreferenceManager(applicationContext)
        notificationHelper = NotificationHelper(applicationContext)
    }

    /**
     * 接收新消息时调用
     */
    override fun onMessageReceived(remoteMessage: RemoteMessage) {
        Log.d(TAG, "From: ${remoteMessage.from}")

        // 处理数据消息
        if (remoteMessage.data.isNotEmpty()) {
            Log.d(TAG, "Message data payload: ${remoteMessage.data}")

            // 从数据消息中提取标题和内容
            val title = remoteMessage.data["title"]
            val message = remoteMessage.data["message"] ?: remoteMessage.data["body"]

            // 显示通知
            notificationHelper.showNotification(title, message)
        }

        // 处理通知消息 (如果应用在前台)
        remoteMessage.notification?.let {
            Log.d(TAG, "Message Notification Title: ${it.title}")
            Log.d(TAG, "Message Notification Body: ${it.body}")

            // 显示通知
            notificationHelper.showNotification(it.title, it.body)
        }
    }

    /**
     * 当新的 Token 生成时调用
     * 场景：
     * 1. 首次安装应用
     * 2. Token 被 invalidated (卸载重装、数据清除等)
     * 3. App Instance 恢复 (如从备份恢复)
     */
    override fun onNewToken(token: String) {
        Log.d(TAG, "Refreshed token: $token")

        // 保存 Token 到本地
        preferenceManager.saveFcmToken(token)

        // 如果用户已登录，发送到服务器
        if (preferenceManager.isLoggedIn()) {
            val jwtToken = preferenceManager.getToken()
            val serverUrl = preferenceManager.getServerUrl()

            if (jwtToken != null && serverUrl != null) {
                sendTokenToServer(serverUrl, jwtToken, token)
            }
        }
    }

    /**
     * 发送 FCM Token 到服务器
     */
    private fun sendTokenToServer(baseUrl: String, jwtToken: String, fcmToken: String) {
        serviceScope.launch {
            try {
                val authService = ApiClient.createAuthorizedClient(baseUrl, jwtToken)
                    .create(AuthService::class.java)

                val request = FcmTokenRequest(
                    token = fcmToken,
                    deviceName = DeviceInfoUtils.getDeviceName()
                )

                val response = authService.submitFcmToken(request)

                if (response.isSuccessful) {
                    Log.d(TAG, "FCM Token sent to server successfully")
                } else {
                    Log.e(TAG, "Failed to send FCM Token: HTTP ${response.code()}")
                }
            } catch (e: Exception) {
                Log.e(TAG, "Error sending FCM Token to server", e)
            }
        }
    }
}
