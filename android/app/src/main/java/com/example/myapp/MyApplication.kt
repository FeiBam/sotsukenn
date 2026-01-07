package com.example.myapp

import android.app.Application
import android.content.Intent
import android.util.Log
import android.widget.Toast
import com.example.myapp.network.ApiClient
import com.example.myapp.network.AuthHandler
import com.example.myapp.utils.PreferenceManager
import com.google.firebase.FirebaseApp
import com.google.firebase.messaging.FirebaseMessaging

class MyApplication : Application(), AuthHandler {

    private lateinit var preferenceManager: PreferenceManager

    companion object {
        private const val TAG = "MyApplication"
    }

    override fun onCreate() {
        super.onCreate()
        preferenceManager = PreferenceManager(applicationContext)

        // Set up global 401 handler
        ApiClient.authHandler = this

        initializeFirebase()
    }

    /**
     * 初始化 Firebase 和 FCM
     */
    private fun initializeFirebase() {
        // Firebase SDK 会自动从 google-services.json 读取配置
        // 但显式初始化更安全
        if (FirebaseApp.getApps(this).isEmpty()) {
            FirebaseApp.initializeApp(this)
            Log.d(TAG, "Firebase initialized successfully")
        }

        // 获取 FCM Token
        getFcmToken()
    }

    /**
     * 获取 FCM Token
     * 注意：首次获取可能需要几秒钟
     */
    private fun getFcmToken() {
        FirebaseMessaging.getInstance().token.addOnCompleteListener { task ->
            if (!task.isSuccessful) {
                Log.w(TAG, "Fetching FCM registration token failed", task.exception)
                return@addOnCompleteListener
            }

            // 获取新的 Token
            val token = task.result
            if (token != null) {
                Log.d(TAG, "FCM Token: $token")

                // 保存到本地
                preferenceManager.saveFcmToken(token)

                // TODO: 如果用户已登录且 Token 未发送，则发送到服务器
                // if (preferenceManager.isLoggedIn() && !preferenceManager.isFcmTokenSentToServer()) {
                //     sendTokenToBackend(token)
                // }
            }
        }
    }

    /**
     * Handle 401 Unauthorized responses
     * Called when API returns 401, indicating the token is invalid or expired
     */
    override fun onUnauthorized() {
        Log.w(TAG, "401 Unauthorized: Clearing token and redirecting to login")

        // Clear the invalid token
        preferenceManager.clearToken()

        // Show message to user
        Toast.makeText(
            applicationContext,
            "Session expired. Please login again.",
            Toast.LENGTH_LONG
        ).show()

        // Redirect to MainActivity (server input/login flow)
        val intent = Intent(this, MainActivity::class.java).apply {
            flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TASK
        }
        startActivity(intent)
    }
}
