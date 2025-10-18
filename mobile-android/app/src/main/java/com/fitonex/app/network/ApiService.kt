package com.fitonex.app.network

import com.fitonex.app.data.model.*
import okhttp3.MediaType.Companion.toMediaType
import retrofit2.Response
import retrofit2.http.*

interface ApiService {
    
    // Auth endpoints
    @POST("v1/auth/register")
    suspend fun register(@Body request: RegisterRequest): Response<AuthResponse>
    
    @POST("v1/auth/login")
    suspend fun login(@Body request: AuthRequest): Response<AuthResponse>
    
    // Gym endpoints
    @GET("v1/gyms/nearby")
    suspend fun getNearbyGyms(
        @Query("lat") lat: Double,
        @Query("lng") lng: Double,
        @Query("radius_km") radiusKm: Double = 5.0,
        @Query("limit") limit: Int = 20,
        @Query("cursor") cursor: String? = null
    ): Response<PaginatedResponse<Gym>>
    
    @GET("v1/gyms/{id}")
    suspend fun getGym(@Path("id") id: String): Response<Gym>
    
    @GET("v1/gyms/{id}/machines")
    suspend fun getGymMachines(@Path("id") id: String): Response<List<Machine>>
    
    @GET("v1/gyms/{id}/prices")
    suspend fun getGymPrices(@Path("id") id: String): Response<List<GymPrice>>
    
    @GET("v1/gyms/{id}/reviews")
    suspend fun getGymReviews(
        @Path("id") id: String,
        @Query("limit") limit: Int = 20,
        @Query("cursor") cursor: String? = null
    ): Response<GymReviewsResponse>
    
    @POST("v1/gyms/{id}/reviews")
    suspend fun createGymReview(
        @Path("id") id: String,
        @Body request: CreateReviewRequest
    ): Response<GymReview>
    
    // Machine endpoints
    @GET("v1/machines")
    suspend fun searchMachines(
        @Query("query") query: String? = null,
        @Query("body_part") bodyPart: String? = null,
        @Query("limit") limit: Int = 20
    ): Response<SearchMachinesResponse>
    
    @GET("v1/machines/body-parts")
    suspend fun getBodyParts(): Response<List<String>>
    
    @GET("v1/machines/{id}")
    suspend fun getMachine(@Path("id") id: String): Response<Machine>
    
    // Video endpoints
    @GET("v1/videos")
    suspend fun getVideos(
        @Query("machine_id") machineId: String,
        @Query("limit") limit: Int = 20,
        @Query("cursor") cursor: String? = null
    ): Response<PaginatedResponse<InstructionVideo>>
    
    @GET("v1/videos/{id}")
    suspend fun getVideo(@Path("id") id: String): Response<InstructionVideo>
    
    @POST("v1/videos/upload-url")
    suspend fun getUploadURL(@Body request: UploadURLRequest): Response<UploadURLResponse>
    
    @POST("v1/videos/finalize")
    suspend fun finalizeVideo(@Body request: FinalizeVideoRequest): Response<InstructionVideo>
    
    @POST("v1/videos/{id}/like")
    suspend fun likeVideo(@Path("id") id: String): Response<Unit>
    
    @DELETE("v1/videos/{id}/like")
    suspend fun unlikeVideo(@Path("id") id: String): Response<Unit>
    
    // Check-in endpoints
    @POST("v1/checkins/today")
    suspend fun checkinToday(): Response<CheckinTodayResponse>
    
    @GET("v1/checkins/me")
    suspend fun getCheckinStats(): Response<CheckinStats>
    
    // Exercise endpoints
    @POST("v1/exercises")
    suspend fun createExercise(@Body request: CreateExerciseRequest): Response<Exercise>

    @GET("v1/exercises")
    suspend fun getExercises(
        @Query("day") day: String,
        @Query("limit") limit: Int = 20,
        @Query("cursor") cursor: String? = null
    ): Response<PaginatedResponse<Exercise>>
    
    @GET("v1/exercises/{id}")
    suspend fun getExercise(@Path("id") id: String): Response<Exercise>
    
    // User endpoints
    @GET("v1/profile")
    suspend fun getProfile(): Response<User>
    
    @PUT("v1/profile")
    suspend fun updateProfile(@Body user: User): Response<User>
    
    companion object {
        fun create(authInterceptor: AuthInterceptor): ApiService {
            val json = kotlinx.serialization.json.Json { ignoreUnknownKeys = true }

            val retrofit = retrofit2.Retrofit.Builder()
                .baseUrl("http://10.0.2.2:8080/") // Android emulator localhost
                .addConverterFactory(json.asConverterFactory("application/json".toMediaType()))
                .client(
                    okhttp3.OkHttpClient.Builder()
                        .addInterceptor(authInterceptor)
                        .addInterceptor(
                            okhttp3.logging.HttpLoggingInterceptor().apply {
                                level = okhttp3.logging.HttpLoggingInterceptor.Level.BODY
                            }
                        )
                        .build()
                )
                .build()
            
            return retrofit.create(ApiService::class.java)
        }
    }
}
