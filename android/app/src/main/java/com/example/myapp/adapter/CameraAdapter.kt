package com.example.myapp.adapter

import android.util.Log
import android.view.LayoutInflater
import android.view.ViewGroup
import androidx.recyclerview.widget.RecyclerView
import coil.request.ImageRequest
import com.example.myapp.databinding.ItemCameraBinding
import com.example.myapp.model.CameraSnapshotResponse
import com.example.myapp.network.ApiClient
import com.example.myapp.network.ApiService
import com.example.myapp.utils.PreferenceManager
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch

class CameraAdapter(
    private val cameras: List<String>,
    private val onCameraClick: (String) -> Unit,
    private val preferenceManager: PreferenceManager
) : RecyclerView.Adapter<CameraAdapter.CameraViewHolder>() {

    inner class CameraViewHolder(private val binding: ItemCameraBinding) :
        RecyclerView.ViewHolder(binding.root) {

        fun bind(cameraName: String) {
            binding.tvCameraName.text = cameraName

            // 加载快照图片
            loadSnapshot(cameraName)

            // 点击事件
            binding.root.setOnClickListener {
                onCameraClick(cameraName)
            }
        }

        private fun loadSnapshot(cameraName: String) {
            val scope = CoroutineScope(Dispatchers.IO)
            scope.launch {
                try {
                    val serverUrl = preferenceManager.getServerUrl()
                    val token = preferenceManager.getToken()

                    if (serverUrl != null && token != null) {
                        val apiService = ApiClient.createAuthorizedClient(serverUrl, token)
                            .create(ApiService::class.java)

                        val response = apiService.getCameraSnapshot(cameraName)

                        if (response.isSuccessful && response.body()?.status == "success") {
                            val snapshotUrl = response.body()!!.body.url
                            val frigateToken = response.body()!!.body.frigateToken

                            // 使用 frigate_token 加载图片
                            loadImageWithToken(snapshotUrl, frigateToken)
                        } else {
                            Log.e("CameraAdapter", "Failed to load snapshot")
                        }
                    }
                } catch (e: Exception) {
                    Log.e("CameraAdapter", "Error loading snapshot", e)
                }
            }
        }

        private fun loadImageWithToken(url: String, frigateToken: String) {
            val request = ImageRequest.Builder(binding.root.context)
                .data(url)
                .crossfade(true)
                .setHeader("Authorization", "Bearer $frigateToken")
                .target(binding.ivSnapshot)
                .build()

            coil.Coil.imageLoader(binding.root.context).enqueue(request)
        }
    }

    override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): CameraViewHolder {
        val binding = ItemCameraBinding.inflate(
            LayoutInflater.from(parent.context),
            parent,
            false
        )
        return CameraViewHolder(binding)
    }

    override fun onBindViewHolder(holder: CameraViewHolder, position: Int) {
        holder.bind(cameras[position])
    }

    override fun getItemCount(): Int = cameras.size
}
