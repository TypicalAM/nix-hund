package com.example.nixhund.screens

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material3.CenterAlignedTopAppBar
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.navigation.NavHostController
import com.example.nixhund.R
import com.example.nixhund.SearchViewModel
import com.example.nixhund.api.ApiClient
import com.example.nixhund.api.HistoryEntry
import com.example.nixhund.getApiKey

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun History(navHostController: NavHostController, searchViewModel: SearchViewModel) {
    var entries by remember { mutableStateOf<List<HistoryEntry>>(listOf()) }
    var isLoading by remember { mutableStateOf(true) }

    val context = LocalContext.current
    val client = ApiClient(getApiKey(context))

    LaunchedEffect(Unit) {
        try {
            entries = client.getHistoryList()
            isLoading = false
        } catch (_: Exception) {
        }
    }

    Scaffold(
        topBar = {
            CenterAlignedTopAppBar(
                title = {
                    Text("History")
                },
                navigationIcon = {
                    IconButton(onClick = {
                        navHostController.navigate("search")
                    }) {
                        Icon(
                            imageVector = Icons.AutoMirrored.Filled.ArrowBack,
                            contentDescription = "Localized description"
                        )
                    }
                },
            )
        },
    ) { contentPadding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(contentPadding),
            verticalArrangement = Arrangement.Top,
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            if (isLoading) {
                CircularProgressIndicator(
                    modifier = Modifier.padding(16.dp),
                )
            } else {
                LazyColumn {
                    itemsIndexed(entries) { _, item ->
                        HistoryListItem(item = item, onClick = {
                            searchViewModel.currentPackage = item.pkg
                            navHostController.navigate("detail")
                        })
                    }
                }
            }
        }
    }
}

@Composable
fun HistoryListItem(item: HistoryEntry, onClick: () -> Unit) {
    Row(modifier = Modifier
        .fillMaxWidth()
        .clickable { onClick() }
        .padding(16.dp),
        verticalAlignment = Alignment.CenterVertically) {
        Icon(
            painter = painterResource(id = R.drawable.nix_snowflake),
            contentDescription = null,
            modifier = Modifier.size(24.dp),
            tint = Color.Gray
        )
        Spacer(modifier = Modifier.width(16.dp))
        Column {
            Text(text = item.pkg.pkgName, style = MaterialTheme.typography.labelSmall)
            Text(text = item.date.toString(), style = MaterialTheme.typography.labelSmall)
            Text(
                text = "Found in ${item.indexID}",
                style = MaterialTheme.typography.labelSmall,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis
            )
        }
    }
}