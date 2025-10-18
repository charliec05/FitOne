package com.fitonex.app.ui.screens

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import com.fitonex.app.data.model.Gym
import com.fitonex.app.data.repository.GymRepository

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun MapScreen() {
    var nearbyGyms by remember { mutableStateOf<List<Gym>>(emptyList()) }
    var isLoading by remember { mutableStateOf(true) }
    var selectedGym by remember { mutableStateOf<Gym?>(null) }
    
    LaunchedEffect(Unit) {
        // TODO: Get nearby gyms from repository
        // For now, show mock data
        nearbyGyms = listOf(
            Gym(
                id = "1",
                name = "FitZone Downtown",
                lat = 47.6097,
                lng = -122.3425,
                address = "1912 Pike Pl, Seattle, WA",
                phone = "+1-206-555-0101",
                website = "https://fitzone.com",
                created_at = "2024-01-01T00:00:00Z",
                distance_m = 250.0,
                avg_rating = 4.5,
                machines_count = 15,
                price_from_cents = 8900
            ),
            Gym(
                id = "2",
                name = "PowerGym Central",
                lat = 47.6133,
                lng = -122.3185,
                address = "1423 10th Ave, Seattle, WA",
                phone = "+1-206-555-0102",
                website = "https://powergym.com",
                created_at = "2024-01-01T00:00:00Z",
                distance_m = 980.0,
                avg_rating = 4.2,
                machines_count = 20,
                price_from_cents = 9900
            )
        )
        isLoading = false
    }
    
    Box(modifier = Modifier.fillMaxSize()) {
        // TODO: Implement Google Maps
        // For now, show a placeholder
        Box(
            modifier = Modifier
                .fillMaxSize()
                .padding(16.dp),
            contentAlignment = Alignment.Center
        ) {
            Card(
                modifier = Modifier.fillMaxWidth(),
                colors = CardDefaults.cardColors(
                    containerColor = MaterialTheme.colorScheme.surfaceVariant
                )
            ) {
                Column(
                    modifier = Modifier.padding(24.dp),
                    horizontalAlignment = Alignment.CenterHorizontally
                ) {
                    Icon(
                        Icons.Filled.Map,
                        contentDescription = "Map",
                        modifier = Modifier.size(64.dp),
                        tint = MaterialTheme.colorScheme.primary
                    )
                    Spacer(modifier = Modifier.height(16.dp))
                    Text(
                        text = "Google Maps Integration",
                        style = MaterialTheme.typography.titleLarge,
                        fontWeight = FontWeight.Bold
                    )
                    Text(
                        text = "Map will be implemented here",
                        style = MaterialTheme.typography.bodyMedium,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
            }
        }
        
        // Bottom sheet with gym list
        if (nearbyGyms.isNotEmpty()) {
            GymListBottomSheet(
                gyms = nearbyGyms,
                selectedGym = selectedGym,
                onGymSelected = { selectedGym = it }
            )
        }
    }
}

@Composable
fun GymListBottomSheet(
    gyms: List<Gym>,
    selectedGym: Gym?,
    onGymSelected: (Gym) -> Unit
) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .padding(16.dp),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surface
        )
    ) {
        Column(
            modifier = Modifier.padding(16.dp)
        ) {
            Text(
                text = "Nearby Gyms",
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.Bold
            )
            
            Spacer(modifier = Modifier.height(8.dp))
            
            LazyColumn(
                verticalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                items(gyms) { gym ->
                    GymListItem(
                        gym = gym,
                        isSelected = selectedGym?.id == gym.id,
                        onClick = { onGymSelected(gym) }
                    )
                }
            }
        }
    }
}

@Composable
fun GymListItem(
    gym: Gym,
    isSelected: Boolean,
    onClick: () -> Unit
) {
    Card(
        onClick = onClick,
        colors = CardDefaults.cardColors(
            containerColor = if (isSelected) {
                MaterialTheme.colorScheme.primaryContainer
            } else {
                MaterialTheme.colorScheme.surface
            }
        )
    ) {
        Column(
            modifier = Modifier.padding(12.dp)
        ) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = gym.name,
                    style = MaterialTheme.typography.titleSmall,
                    fontWeight = FontWeight.Bold
                )
                
                gym.avg_rating?.let { rating ->
                    Row(
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Icon(
                            Icons.Filled.Star,
                            contentDescription = "Rating",
                            modifier = Modifier.size(16.dp),
                            tint = MaterialTheme.colorScheme.primary
                        )
                        Spacer(modifier = Modifier.width(4.dp))
                        Text(
                            text = String.format("%.1f", rating),
                            style = MaterialTheme.typography.bodySmall
                        )
                    }
                }
            }
            
            Text(
                text = gym.address,
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
            
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                gym.distance_m?.let { distance ->
                    Text(
                        text = "${String.format("%.1f", distance / 1000)} km",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.primary
                    )
                }
                
                Text(
                    text = "${gym.machines_count} machines",
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }
        }
    }
}
