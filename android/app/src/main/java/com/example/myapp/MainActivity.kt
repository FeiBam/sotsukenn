package com.example.myapp

import android.content.Intent
import android.os.Bundle
import android.view.View
import android.widget.Button
import android.widget.ProgressBar
import android.widget.TextView
import android.widget.Toast
import androidx.appcompat.app.AppCompatActivity
import androidx.lifecycle.ViewModelProvider
import com.example.myapp.utils.PreferenceManager
import com.example.myapp.viewmodel.ServerState
import com.example.myapp.viewmodel.ServerViewModel
import com.google.android.material.textfield.TextInputEditText

class MainActivity : AppCompatActivity() {

    private lateinit var viewModel: ServerViewModel
    private lateinit var preferenceManager: PreferenceManager
    private var isResumingSession = false

    private lateinit var etServerUrl: TextInputEditText
    private lateinit var btnConnect: Button
    private lateinit var progressBar: ProgressBar
    private lateinit var tvError: TextView

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_main)

        preferenceManager = PreferenceManager(this)
        viewModel = ViewModelProvider(this)[ServerViewModel::class.java]

        initViews()
        observeViewModel()

        // Check if user has a previous session
        val serverUrl = preferenceManager.getServerUrl()
        if (preferenceManager.isLoggedIn() && serverUrl != null) {
            // Attempt to resume session by validating server
            isResumingSession = true
            etServerUrl.setText(serverUrl)
            attemptResumeSession(serverUrl)
        } else {
            // First-time setup or logged out
            isResumingSession = false
            loadSavedServerUrl()
        }
    }

    private fun initViews() {
        etServerUrl = findViewById(R.id.etServerUrl)
        btnConnect = findViewById(R.id.btnConnect)
        progressBar = findViewById(R.id.progressBar)
        tvError = findViewById(R.id.tvError)

        btnConnect.setOnClickListener {
            val serverUrl = etServerUrl.text?.toString()?.trim()
            if (serverUrl.isNullOrEmpty()) {
                tvError.text = "Please enter a server URL"
                tvError.visibility = View.VISIBLE
            } else {
                tvError.visibility = View.GONE
                viewModel.validateServer(serverUrl)
            }
        }
    }

    private fun observeViewModel() {
        viewModel.state.observe(this) { state ->
            when (state) {
                is ServerState.Idle -> {
                    // Do nothing
                }
                is ServerState.Loading -> {
                    btnConnect.isEnabled = false
                    progressBar.visibility = View.VISIBLE
                    tvError.visibility = View.GONE
                }
                is ServerState.Success -> {
                    val serverUrl = etServerUrl.text?.toString()?.trim() ?: return@observe
                    preferenceManager.saveServerUrl(serverUrl)

                    btnConnect.isEnabled = true
                    progressBar.visibility = View.GONE
                    etServerUrl.isEnabled = true

                    if (isResumingSession) {
                        // Silent navigation to HomeActivity
                        navigateToHome()
                    } else {
                        // First-time setup, show success message
                        Toast.makeText(this, "Server connected!", Toast.LENGTH_SHORT).show()

                        // Navigate to LoginActivity
                        val intent = Intent(this, LoginActivity::class.java)
                        startActivity(intent)
                    }
                }
                is ServerState.Error -> {
                    btnConnect.isEnabled = true
                    progressBar.visibility = View.GONE
                    etServerUrl.isEnabled = true
                    tvError.text = state.message
                    tvError.visibility = View.VISIBLE

                    // If we were trying to resume session but failed, let user know they can update URL
                    if (isResumingSession) {
                        isResumingSession = false
                        tvError.text = "Cannot connect to server. Please check the server address or your network connection."
                    }
                }
            }
        }
    }

    private fun loadSavedServerUrl() {
        preferenceManager.getServerUrl()?.let {
            etServerUrl.setText(it)
        }
    }

    private fun attemptResumeSession(serverUrl: String) {
        progressBar.visibility = View.VISIBLE
        tvError.visibility = View.GONE
        btnConnect.isEnabled = false
        etServerUrl.isEnabled = false

        viewModel.validateServer(serverUrl)
    }

    private fun navigateToHome() {
        val intent = Intent(this, HomeActivity::class.java)
        startActivity(intent)
        finish()
    }
}
