package com.example.myapp

import android.os.Bundle
import android.view.View
import android.webkit.WebChromeClient
import android.webkit.WebView
import android.webkit.WebViewClient
import android.widget.Toast
import androidx.appcompat.app.AppCompatActivity
import androidx.lifecycle.ViewModelProvider
import com.example.myapp.utils.PreferenceManager
import com.example.myapp.viewmodel.CameraState
import com.example.myapp.viewmodel.CameraViewModel


class CameraPlayerActivity : AppCompatActivity() {

    private lateinit var viewModel: CameraViewModel
    private lateinit var preferenceManager: PreferenceManager
    private lateinit var webView: WebView
    private lateinit var progressBar: View

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_camera_player)

        val cameraName = intent.getStringExtra("camera_name") ?: run {
            finish()
            return
        }

        preferenceManager = PreferenceManager(this)
        viewModel = ViewModelProvider(this)[CameraViewModel::class.java]
        viewModel.setPreferenceManager(preferenceManager)

        setupToolbar(cameraName)
        initViews()
        observeViewModel()
        viewModel.loadCameraStream(cameraName)
    }

    private fun setupToolbar(cameraName: String) {
        val toolbar = findViewById<androidx.appcompat.widget.Toolbar>(R.id.toolbar)
        toolbar.title = cameraName
        toolbar.setNavigationOnClickListener {
            finish()
        }
    }

    private fun initViews() {
        webView = findViewById(R.id.webView)
        progressBar = findViewById(R.id.progressBar)

        setupWebView()
    }

    private fun setupWebView() {
        webView.settings.apply {
            javaScriptEnabled = true
            domStorageEnabled = true
            loadWithOverviewMode = true
            useWideViewPort = true
        }

        webView.webViewClient = object : WebViewClient() {
            override fun onPageFinished(view: WebView?, url: String?) {
                super.onPageFinished(view, url)
                progressBar.visibility = View.GONE
            }

            override fun onReceivedError(
                view: WebView?,
                errorCode: Int,
                description: String?,
                failingUrl: String?
            ) {
                super.onReceivedError(view, errorCode, description, failingUrl)
                progressBar.visibility = View.GONE
                Toast.makeText(
                    this@CameraPlayerActivity,
                    "Error loading stream: $description",
                    Toast.LENGTH_LONG
                ).show()
            }
        }

        webView.webChromeClient = WebChromeClient()

    }

    private fun observeViewModel() {
        viewModel.state.observe(this) { state ->
            when (state) {
                is CameraState.Loading -> {
                    progressBar.visibility = View.VISIBLE
                }
                is CameraState.StreamLoaded -> {
                    loadStream(state.streamUrl, state.token)
                }
                is CameraState.Error -> {
                    progressBar.visibility = View.GONE
                    Toast.makeText(this, state.message, Toast.LENGTH_LONG).show()
                }
                else -> {}
            }
        }
    }

    private fun loadStream(url: String, token: String) {
        val headers: MutableMap<String?, String?> = HashMap<String?, String?>()
        headers["Authorization"] = "Bearer $token"
        webView.loadUrl(url,headers)
    }

    override fun onBackPressed() {
        if (webView.canGoBack()) {
            webView.goBack()
        } else {
            super.onBackPressed()
        }
    }

    override fun onDestroy() {
        super.onDestroy()
        // 清理 WebView 以防止内存泄漏
        webView.destroy()
    }
}
