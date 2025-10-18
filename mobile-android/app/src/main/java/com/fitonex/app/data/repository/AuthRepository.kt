package com.fitonex.app.data.repository

import android.content.Context
import android.content.SharedPreferences
import com.fitonex.app.data.model.AuthRequest
import com.fitonex.app.data.model.AuthResponse
import com.fitonex.app.data.model.RegisterRequest
import com.fitonex.app.network.ApiService
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

class AuthRepository(
    private val apiService: ApiService,
    private val context: Context
) {
    private val prefs: SharedPreferences = context.getSharedPreferences("auth", Context.MODE_PRIVATE)
    
    private val _isLoggedIn = MutableStateFlow(getToken() != null)
    val isLoggedIn: StateFlow<Boolean> = _isLoggedIn.asStateFlow()
    
    fun getToken(): String? {
        return prefs.getString("token", null)
    }
    
    private fun saveToken(token: String) {
        prefs.edit().putString("token", token).apply()
        _isLoggedIn.value = true
    }
    
    private fun clearToken() {
        prefs.edit().remove("token").apply()
        _isLoggedIn.value = false
    }
    
    suspend fun login(email: String, password: String): Result<AuthResponse> {
        return try {
            val response = apiService.login(AuthRequest(email, password))
            if (response.isSuccessful) {
                val authResponse = response.body()!!
                saveToken(authResponse.token)
                Result.success(authResponse)
            } else {
                Result.failure(Exception("Login failed: ${response.message()}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    suspend fun register(email: String, password: String, name: String): Result<AuthResponse> {
        return try {
            val response = apiService.register(RegisterRequest(email, password, name))
            if (response.isSuccessful) {
                val authResponse = response.body()!!
                saveToken(authResponse.token)
                Result.success(authResponse)
            } else {
                Result.failure(Exception("Registration failed: ${response.message()}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    fun logout() {
        clearToken()
    }
}
