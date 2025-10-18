package com.fitonex.app.data.repository

import com.fitonex.app.data.model.*
import com.fitonex.app.network.ApiService

class VideoRepository(
    private val apiService: ApiService
) {
    
    suspend fun getVideos(
        machineId: String,
        limit: Int = 20,
        cursor: String? = null
    ): Result<PaginatedResponse<InstructionVideo>> {
        return try {
            val response = apiService.getVideos(machineId, limit, cursor)
            if (response.isSuccessful) {
                Result.success(response.body()!!)
            } else {
                Result.failure(Exception("Failed to fetch videos: ${response.message()}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    suspend fun getVideo(id: String): Result<InstructionVideo> {
        return try {
            val response = apiService.getVideo(id)
            if (response.isSuccessful) {
                Result.success(response.body()!!)
            } else {
                Result.failure(Exception("Failed to fetch video: ${response.message()}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    suspend fun getUploadURL(
        machineId: String,
        title: String,
        description: String,
        contentType: String,
        bytes: Long
    ): Result<UploadURLResponse> {
        return try {
            val response = apiService.getUploadURL(
                UploadURLRequest(machineId, title, description, contentType, bytes)
            )
            if (response.isSuccessful) {
                Result.success(response.body()!!)
            } else {
                Result.failure(Exception("Failed to get upload URL: ${response.message()}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    suspend fun finalizeVideo(
        machineId: String,
        title: String,
        description: String,
        videoKey: String,
        thumbKey: String? = null,
        durationSec: Int
    ): Result<InstructionVideo> {
        return try {
            val response = apiService.finalizeVideo(
                FinalizeVideoRequest(machineId, title, description, videoKey, thumbKey, durationSec)
            )
            if (response.isSuccessful) {
                Result.success(response.body()!!)
            } else {
                Result.failure(Exception("Failed to finalize video: ${response.message()}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    suspend fun likeVideo(id: String): Result<Unit> {
        return try {
            val response = apiService.likeVideo(id)
            if (response.isSuccessful) {
                Result.success(Unit)
            } else {
                Result.failure(Exception("Failed to like video: ${response.message()}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    suspend fun unlikeVideo(id: String): Result<Unit> {
        return try {
            val response = apiService.unlikeVideo(id)
            if (response.isSuccessful) {
                Result.success(Unit)
            } else {
                Result.failure(Exception("Failed to unlike video: ${response.message()}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
}
