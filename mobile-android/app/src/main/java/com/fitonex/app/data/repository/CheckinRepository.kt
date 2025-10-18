package com.fitonex.app.data.repository

import com.fitonex.app.data.model.CheckinStats
import com.fitonex.app.data.model.CheckinTodayResponse
import com.fitonex.app.network.ApiService

class CheckinRepository(
    private val apiService: ApiService
) {
    
    suspend fun checkinToday(): Result<CheckinTodayResponse> {
        return try {
            val response = apiService.checkinToday()
            if (response.isSuccessful) {
                Result.success(response.body()!!)
            } else {
                Result.failure(Exception("Failed to check in: ${response.message()}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    suspend fun getCheckinStats(): Result<CheckinStats> {
        return try {
            val response = apiService.getCheckinStats()
            if (response.isSuccessful) {
                Result.success(response.body()!!)
            } else {
                Result.failure(Exception("Failed to get check-in stats: ${response.message()}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
}
