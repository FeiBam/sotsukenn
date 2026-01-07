package com.example.myapp.utils

import android.os.Build

/**
 * 设备信息工具类
 */
object DeviceInfoUtils {

    /**
     * 获取设备名称
     * 格式：制造商 + 型号（例如：Samsung Galaxy S21）
     */
    fun getDeviceName(): String {
        val manufacturer = Build.MANUFACTURER
        val model = Build.MODEL

        return if (model.startsWith(manufacturer)) {
            model
        } else {
            "$manufacturer $model"
        }
    }

    /**
     * 获取 Android 版本
     */
    fun getAndroidVersion(): String {
        return "Android ${Build.VERSION.RELEASE}"
    }
}
