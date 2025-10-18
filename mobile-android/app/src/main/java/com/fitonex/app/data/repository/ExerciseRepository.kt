package com.fitonex.app.data.repository

import com.fitonex.app.data.model.*
import com.fitonex.app.network.ApiService

class ExerciseRepository(
    private val apiService: ApiService
) {
    
    suspend fun createExercise(
        day: String,
        gymId: String? = null,
        machineId: String? = null,
        name: String,
        sets: List<Set>
    ): Result<Exercise> {
        return try {
            val response = apiService.createExercise(
                CreateExerciseRequest(day, gymId, machineId, name, sets)
            )
            if (response.isSuccessful) {
                Result.success(response.body()!!)
            } else {
                Result.failure(Exception("Failed to create exercise: ${response.message()}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun getExercises(
        day: String,
        limit: Int = 20,
        cursor: String? = null
    ): Result<PaginatedResponse<Exercise>> {
        return try {
            val response = apiService.getExercises(day, limit, cursor)
            if (response.isSuccessful) {
                Result.success(response.body()!!)
            } else {
                Result.failure(Exception("Failed to fetch exercises: ${response.message()}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    suspend fun getExercise(id: String): Result<Exercise> {
        return try {
            val response = apiService.getExercise(id)
            if (response.isSuccessful) {
                Result.success(response.body()!!)
            } else {
                Result.failure(Exception("Failed to fetch exercise: ${response.message()}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
}
