package com.fitonex.app.ui.screens

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import com.fitonex.app.data.model.InstructionVideo
import com.fitonex.app.data.model.Machine
import com.fitonex.app.data.repository.VideoRepository

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun LearnScreen() {
    var machines by remember { mutableStateOf<List<Machine>>(emptyList()) }
    var selectedMachine by remember { mutableStateOf<Machine?>(null) }
    var videos by remember { mutableStateOf<List<InstructionVideo>>(emptyList()) }
    var isLoading by remember { mutableStateOf(true) }
    
    LaunchedEffect(Unit) {
        // TODO: Get machines from repository
        machines = listOf(
            Machine("1", "Bench Press", "Chest", "2024-01-01T00:00:00Z"),
            Machine("2", "Squat Rack", "Legs", "2024-01-01T00:00:00Z"),
            Machine("3", "Deadlift Platform", "Back", "2024-01-01T00:00:00Z"),
            Machine("4", "Pull-up Bar", "Back", "2024-01-01T00:00:00Z"),
            Machine("5", "Dumbbells", "Arms", "2024-01-01T00:00:00Z")
        )
        isLoading = false
    }
    
    LazyColumn(
        modifier = Modifier
            .fillMaxSize()
            .padding(16.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        item {
            Text(
                text = "Learn",
                style = MaterialTheme.typography.headlineMedium,
                fontWeight = FontWeight.Bold
            )
        }
        
        item {
            Text(
                text = "Machines",
                style = MaterialTheme.typography.titleLarge,
                fontWeight = FontWeight.Bold
            )
        }
        
        item {
            LazyRow(
                horizontalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                items(machines) { machine ->
                    MachineChip(
                        machine = machine,
                        isSelected = selectedMachine?.id == machine.id,
                        onClick = { selectedMachine = machine }
                    )
                }
            }
        }
        
        if (selectedMachine != null) {
            item {
                Text(
                    text = "${selectedMachine!!.name} Videos",
                    style = MaterialTheme.typography.titleLarge,
                    fontWeight = FontWeight.Bold
                )
            }
            
            if (videos.isEmpty()) {
                item {
                    EmptyVideosCard(
                        machineName = selectedMachine!!.name,
                        onUploadVideo = { /* TODO: Navigate to upload */ }
                    )
                }
            } else {
                items(videos) { video ->
                    VideoCard(video = video)
                }
            }
        } else {
            item {
                Card(
                    modifier = Modifier.fillMaxWidth()
                ) {
                    Column(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(24.dp),
                        horizontalAlignment = Alignment.CenterHorizontally
                    ) {
                        Icon(
                            Icons.Filled.School,
                            contentDescription = "Learn",
                            modifier = Modifier.size(48.dp),
                            tint = MaterialTheme.colorScheme.primary
                        )
                        
                        Spacer(modifier = Modifier.height(16.dp))
                        
                        Text(
                            text = "Select a machine to view instruction videos",
                            style = MaterialTheme.typography.titleMedium,
                            fontWeight = FontWeight.Bold
                        )
                        
                        Text(
                            text = "Learn proper form and technique from expert trainers",
                            style = MaterialTheme.typography.bodyMedium,
                            color = MaterialTheme.colorScheme.onSurfaceVariant
                        )
                    }
                }
            }
        }
    }
}

@Composable
fun MachineChip(
    machine: Machine,
    isSelected: Boolean,
    onClick: () -> Unit
) {
    FilterChip(
        onClick = onClick,
        label = { Text(machine.name) },
        selected = isSelected,
        leadingIcon = {
            Icon(
                Icons.Filled.FitnessCenter,
                contentDescription = null,
                modifier = Modifier.size(18.dp)
            )
        }
    )
}

@Composable
fun EmptyVideosCard(
    machineName: String,
    onUploadVideo: () -> Unit
) {
    Card(
        modifier = Modifier.fillMaxWidth()
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(24.dp),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Icon(
                Icons.Filled.VideoLibrary,
                contentDescription = "No videos",
                modifier = Modifier.size(48.dp),
                tint = MaterialTheme.colorScheme.onSurfaceVariant
            )
            
            Spacer(modifier = Modifier.height(16.dp))
            
            Text(
                text = "No videos for $machineName",
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.Bold
            )
            
            Text(
                text = "Be the first to upload an instruction video!",
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
            
            Spacer(modifier = Modifier.height(16.dp))
            
            Button(onClick = onUploadVideo) {
                Icon(Icons.Filled.Upload, contentDescription = null)
                Spacer(modifier = Modifier.width(8.dp))
                Text("Upload Video")
            }
        }
    }
}

@Composable
fun VideoCard(video: InstructionVideo) {
    Card(
        modifier = Modifier.fillMaxWidth()
    ) {
        Column(
            modifier = Modifier.padding(16.dp)
        ) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = video.title,
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.Bold,
                    modifier = Modifier.weight(1f)
                )
                
                Row(
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Icon(
                        Icons.Filled.Favorite,
                        contentDescription = "Likes",
                        modifier = Modifier.size(16.dp),
                        tint = MaterialTheme.colorScheme.primary
                    )
                    Spacer(modifier = Modifier.width(4.dp))
                    Text(
                        text = "${video.like_count}",
                        style = MaterialTheme.typography.bodySmall
                    )
                }
            }
            
            video.description?.let { description ->
                Spacer(modifier = Modifier.height(4.dp))
                Text(
                    text = description,
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }
            
            Spacer(modifier = Modifier.height(8.dp))
            
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                video.uploader_name?.let { uploaderName ->
                    Text(
                        text = "by $uploaderName",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
                
                video.duration_sec?.let { duration ->
                    Text(
                        text = "${duration / 60}:${String.format("%02d", duration % 60)}",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
            }
            
            Spacer(modifier = Modifier.height(12.dp))
            
            Row(
                horizontalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                Button(
                    onClick = { /* TODO: Play video */ },
                    modifier = Modifier.weight(1f)
                ) {
                    Icon(Icons.Filled.PlayArrow, contentDescription = null)
                    Spacer(modifier = Modifier.width(4.dp))
                    Text("Watch")
                }
                
                IconButton(
                    onClick = { /* TODO: Toggle like */ }
                ) {
                    Icon(
                        if (video.is_liked) Icons.Filled.Favorite else Icons.Filled.FavoriteBorder,
                        contentDescription = if (video.is_liked) "Unlike" else "Like",
                        tint = if (video.is_liked) {
                            MaterialTheme.colorScheme.primary
                        } else {
                            MaterialTheme.colorScheme.onSurfaceVariant
                        }
                    )
                }
            }
        }
    }
}
