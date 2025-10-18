package com.fitonex.app.data.repository

import com.fitonex.app.data.model.*
import com.fitonex.app.network.ApiService

class GymRepository(
    private val apiService: ApiService
) {
    
    suspend fun getNearbyGyms(
        lat: Double,
        lng: Double,
        radiusKm: Double = 5.0,
        limit: Int = 20,
        cursor: String? = null
    ): Result<PaginatedResponse<Gym>> {
        return try {
            val response = apiService.getNearbyGyms(lat, lng, radiusKm, limit, cursor)
            if (response.isSuccessful) {
                Result.success(response.body()!!)
            } else {
                Result.failure(Exception("Failed to fetch nearby gyms: ${response.message()}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    suspend fun getGym(id: String): Result<Gym> {
        return try {
            val response = apiService.getGym(id)
            if (response.isSuccessful) {
                Result.success(response.body()!!)
            } else {
                Result.failure(Exception("Failed to fetch gym: ${response.message()}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    suspend fun getGymMachines(id: String): Result<List<Machine>> {
        return try {
            val response = apiService.getGymMachines(id)
            if (response.isSuccessful) {
                Result.success(response.body()!!)
            } else {
                Result.failure(Exception("Failed to fetch gym machines: ${response.message()}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    suspend fun getGymPrices(id: String): Result<List<GymPrice>> {
        return try {
            val response = apiService.getGymPrices(id)
            if (response.isSuccessful) {
                Result.success(response.body()!!)
            } else {
                Result.failure(Exception("Failed to fetch gym prices: ${response.message()}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    suspend fun getGymReviews(
        id: String,
        limit: Int = 20,
        cursor: String? = null
    ): Result<GymReviewsResponse> {
        return try {
            val response = apiService.getGymReviews(id, limit, cursor)
            if (response.isSuccessful) {
                Result.success(response.body()!!)
            } else {
                Result.failure(Exception("Failed to fetch gym reviews: ${response.message()}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    suspend fun createGymReview(
        gymId: String,
        rating: Int,
        comment: String
    ): Result<GymReview> {
        return try {
            val response = apiService.createGymReview(gymId, CreateReviewRequest(rating, comment))
            if (response.isSuccessful) {
                Result.success(response.body()!!)
            } else {
                Result.failure(Exception("Failed to create review: ${response.message()}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
}
