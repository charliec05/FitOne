package com.fitonex.app

import android.app.Application
import com.fitonex.app.data.repository.AuthRepository
import com.fitonex.app.data.repository.GymRepository
import com.fitonex.app.data.repository.VideoRepository
import com.fitonex.app.data.repository.CheckinRepository
import com.fitonex.app.data.repository.ExerciseRepository
import com.fitonex.app.network.ApiService
import com.fitonex.app.network.AuthInterceptor

class FitONEXApplication : Application() {
    
    // Repositories
    val authRepository by lazy { AuthRepository(apiService) }
    val gymRepository by lazy { GymRepository(apiService) }
    val videoRepository by lazy { VideoRepository(apiService) }
    val checkinRepository by lazy { CheckinRepository(apiService) }
    val exerciseRepository by lazy { ExerciseRepository(apiService) }
    
    // API Service
    private val apiService by lazy { 
        ApiService.create(authInterceptor = AuthInterceptor(authRepository))
    }
}
