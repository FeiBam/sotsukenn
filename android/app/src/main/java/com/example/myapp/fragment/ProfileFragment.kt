package com.example.myapp.fragment

import android.content.Intent
import android.os.Bundle
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import androidx.fragment.app.Fragment
import com.example.myapp.MainActivity
import com.example.myapp.R
import com.example.myapp.utils.PreferenceManager

class ProfileFragment : Fragment() {

    private lateinit var preferenceManager: PreferenceManager

    override fun onCreateView(
        inflater: LayoutInflater,
        container: ViewGroup?,
        savedInstanceState: Bundle?
    ): View? {
        return inflater.inflate(R.layout.fragment_profile, container, false)
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)

        preferenceManager = PreferenceManager(requireContext())

        initViews(view)
        displayUserInfo()
    }

    private fun initViews(view: View) {
        val btnLogout = view.findViewById<com.google.android.material.button.MaterialButton>(R.id.btnLogout)
        btnLogout.setOnClickListener {
            logout()
        }
    }

    private fun displayUserInfo() {
        val username = preferenceManager.getUsername()
        val userInfo = preferenceManager.getUserInfo()

        view?.findViewById<com.google.android.material.textview.MaterialTextView>(R.id.tvUsername)?.text =
            "Username: $username"
        view?.findViewById<com.google.android.material.textview.MaterialTextView>(R.id.tvEmail)?.text =
            "Email: ${userInfo?.email ?: "N/A"}"
        view?.findViewById<com.google.android.material.textview.MaterialTextView>(R.id.tvUserId)?.text =
            "User ID: ${userInfo?.id ?: "N/A"}"
    }

    private fun logout() {
        preferenceManager.clearToken()
        // 返回到登录页面
        val intent = Intent(requireContext(), MainActivity::class.java)
        intent.flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TASK
        startActivity(intent)
        requireActivity().finish()
    }

    companion object {
        fun newInstance() = ProfileFragment()
    }
}
