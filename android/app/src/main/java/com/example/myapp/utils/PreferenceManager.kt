package com.example.myapp.utils

import android.content.Context
import android.content.SharedPreferences
import com.example.myapp.model.UserInfo
import com.google.gson.Gson

class PreferenceManager(context: Context) {

    private val prefs: SharedPreferences = context.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE)
    private val gson = Gson()

    companion object {
        private const val PREFS_NAME = "MyAppPrefs"
        private const val KEY_SERVER_URL = "server_url"
        private const val KEY_TOKEN = "auth_token"
        private const val KEY_USERNAME = "username"
        private const val KEY_USER_INFO = "user_info"
        private const val KEY_FCM_TOKEN = "fcm_token"
    }

    fun saveServerUrl(url: String) {
        prefs.edit().putString(KEY_SERVER_URL, url).apply()
    }

    fun getServerUrl(): String? {
        return prefs.getString(KEY_SERVER_URL, null)
    }

    fun saveToken(token: String) {
        prefs.edit().putString(KEY_TOKEN, token).apply()
    }

    fun getToken(): String? {
        return prefs.getString(KEY_TOKEN, null)
    }

    fun clearToken() {
        prefs.edit().remove(KEY_TOKEN).apply()
    }

    fun saveUsername(username: String) {
        prefs.edit().putString(KEY_USERNAME, username).apply()
    }

    fun getUsername(): String? {
        return prefs.getString(KEY_USERNAME, null)
    }

    fun clearAll() {
        prefs.edit().clear().apply()
    }

    fun saveUserInfo(userInfo: UserInfo) {
        val json = gson.toJson(userInfo)
        prefs.edit().putString(KEY_USER_INFO, json).apply()
    }

    fun getUserInfo(): UserInfo? {
        val json = prefs.getString(KEY_USER_INFO, null)
        return if (json != null) {
            gson.fromJson(json, UserInfo::class.java)
        } else {
            null
        }
    }

    fun isLoggedIn(): Boolean {
        return getToken() != null
    }

    // ========== FCM Token 相关方法 ==========

    /**
     * 保存 FCM Token
     */
    fun saveFcmToken(token: String) {
        prefs.edit().putString(KEY_FCM_TOKEN, token).apply()
    }

    /**
     * 获取 FCM Token
     */
    fun getFcmToken(): String? {
        return prefs.getString(KEY_FCM_TOKEN, null)
    }

    /**
     * 清除 FCM Token (退出登录时使用)
     */
    fun clearFcmToken() {
        prefs.edit().remove(KEY_FCM_TOKEN).apply()
    }
}
