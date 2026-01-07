package com.example.myapp.fragment

import android.content.Intent
import android.os.Bundle
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.Toast
import androidx.fragment.app.Fragment
import androidx.lifecycle.ViewModelProvider
import androidx.recyclerview.widget.LinearLayoutManager
import androidx.recyclerview.widget.RecyclerView
import com.example.myapp.CameraPlayerActivity
import com.example.myapp.R
import com.example.myapp.adapter.CameraAdapter
import com.example.myapp.utils.PreferenceManager
import com.example.myapp.viewmodel.CameraState
import com.example.myapp.viewmodel.CameraViewModel

class CamerasFragment : Fragment() {

    private lateinit var viewModel: CameraViewModel
    private lateinit var preferenceManager: PreferenceManager
    private lateinit var rvCameras: RecyclerView
    private lateinit var progressBar: View
    private var adapter: CameraAdapter? = null

    override fun onCreateView(
        inflater: LayoutInflater,
        container: ViewGroup?,
        savedInstanceState: Bundle?
    ): View? {
        return inflater.inflate(R.layout.fragment_cameras, container, false)
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)

        preferenceManager = PreferenceManager(requireContext())
        viewModel = ViewModelProvider(this)[CameraViewModel::class.java]
        viewModel.setPreferenceManager(preferenceManager)

        initViews(view)
        observeViewModel()
        viewModel.loadCameras()
    }

    private fun initViews(view: View) {
        rvCameras = view.findViewById(R.id.rvCameras)
        progressBar = view.findViewById(R.id.progressBar)

        rvCameras.layoutManager = LinearLayoutManager(requireContext())
    }

    private fun observeViewModel() {
        viewModel.state.observe(viewLifecycleOwner) { state ->
            when (state) {
                is CameraState.Loading -> {
                    progressBar.visibility = View.VISIBLE
                }
                is CameraState.CamerasLoaded -> {
                    progressBar.visibility = View.GONE
                    setupAdapter(state.cameras)
                }
                is CameraState.Error -> {
                    progressBar.visibility = View.GONE
                    Toast.makeText(requireContext(), state.message, Toast.LENGTH_LONG).show()
                }
                else -> {}
            }
        }
    }

    private fun setupAdapter(cameras: List<String>) {
        adapter = CameraAdapter(
            cameras = cameras,
            onCameraClick = { cameraName ->
                openCameraPlayer(cameraName)
            },
            preferenceManager = preferenceManager
        )
        rvCameras.adapter = adapter
    }

    private fun openCameraPlayer(cameraName: String) {
        val intent = Intent(requireContext(), CameraPlayerActivity::class.java)
        intent.putExtra("camera_name", cameraName)
        startActivity(intent)
    }

    companion object {
        fun newInstance() = CamerasFragment()
    }
}
