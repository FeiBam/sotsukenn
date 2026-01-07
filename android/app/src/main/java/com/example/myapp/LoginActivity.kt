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
import com.example.myapp.viewmodel.LoginState
import com.example.myapp.viewmodel.LoginViewModel
import com.google.android.material.textfield.TextInputEditText

class LoginActivity : AppCompatActivity() {

    private lateinit var viewModel: LoginViewModel
    private lateinit var preferenceManager: PreferenceManager

    private lateinit var etUsername: TextInputEditText
    private lateinit var etPassword: TextInputEditText
    private lateinit var btnLogin: Button
    private lateinit var progressBar: ProgressBar
    private lateinit var tvError: TextView

    private var serverUrl: String? = null

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_login)

        preferenceManager = PreferenceManager(this)
        viewModel = ViewModelProvider(this)[LoginViewModel::class.java]

        // 设置 PreferenceManager 到 ViewModel
        viewModel.setPreferenceManager(preferenceManager)

        serverUrl = preferenceManager.getServerUrl()

        if (serverUrl == null) {
            Toast.makeText(this, "Server URL not found. Please connect first.", Toast.LENGTH_LONG).show()
            finish()
            return
        }

        initViews()
        observeViewModel()
    }

    private fun initViews() {
        etUsername = findViewById(R.id.etUsername)
        etPassword = findViewById(R.id.etPassword)
        btnLogin = findViewById(R.id.btnLogin)
        progressBar = findViewById(R.id.progressBar)
        tvError = findViewById(R.id.tvError)

        // Load saved username
        preferenceManager.getUsername()?.let {
            etUsername.setText(it)
        }

        btnLogin.setOnClickListener {
            val username = etUsername.text?.toString()?.trim()
            val password = etPassword.text?.toString()?.trim()

            when {
                username.isNullOrEmpty() -> {
                    tvError.text = "Please enter username"
                    tvError.visibility = View.VISIBLE
                }
                password.isNullOrEmpty() -> {
                    tvError.text = "Please enter password"
                    tvError.visibility = View.VISIBLE
                }
                else -> {
                    tvError.visibility = View.GONE
                    viewModel.login(serverUrl!!, username, password)
                }
            }
        }
    }

    private fun observeViewModel() {
        viewModel.state.observe(this) { state ->
            when (state) {
                is LoginState.Idle -> {
                    // Do nothing
                }
                is LoginState.Loading -> {
                    btnLogin.isEnabled = false
                    progressBar.visibility = View.VISIBLE
                    tvError.visibility = View.GONE
                }
                is LoginState.Success -> {
                    // Save login info
                    preferenceManager.saveToken(state.token)
                    preferenceManager.saveUsername(state.username)
                    state.user?.let { preferenceManager.saveUserInfo(it) }

                    btnLogin.isEnabled = true
                    progressBar.visibility = View.GONE

                    Toast.makeText(this, "Login successful!", Toast.LENGTH_SHORT).show()

                    // Navigate to HomeActivity
                    val intent = Intent(this, HomeActivity::class.java)
                    startActivity(intent)
                    finish()
                }
                is LoginState.Error -> {
                    btnLogin.isEnabled = true
                    progressBar.visibility = View.GONE
                    tvError.text = state.message
                    tvError.visibility = View.VISIBLE
                }
            }
        }
    }
}
