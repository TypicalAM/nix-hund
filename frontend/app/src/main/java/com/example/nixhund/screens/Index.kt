package com.example.nixhund.screens

import android.annotation.SuppressLint
import android.util.Log
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.itemsIndexed

import androidx.compose.foundation.text.BasicTextField
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.material.icons.filled.Search
import androidx.compose.material3.Button
import androidx.compose.material3.CenterAlignedTopAppBar
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.RadioButton
import androidx.compose.material3.Scaffold
import androidx.compose.material3.SnackbarHost
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.Text
import androidx.compose.material3.TextField
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.material3.rememberTopAppBarState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.RectangleShape
import androidx.compose.ui.input.nestedscroll.nestedScroll
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp

import androidx.navigation.NavHostController
import com.example.nixhund.API_KEY
import com.example.nixhund.LOGGED_IN
import com.example.nixhund.SearchViewModel
import com.example.nixhund.api.LoginClient
import com.example.nixhund.api.LoginInfo
import com.example.nixhund.USERNAME
import com.example.nixhund.api.ApiClient
import com.example.nixhund.api.IndexInfo
import com.example.nixhund.getApiKey
import com.example.nixhund.getPreferenceString
import com.example.nixhund.setPref
import kotlinx.coroutines.cancel
import kotlinx.coroutines.launch
import java.text.SimpleDateFormat

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun Index(navHostController: NavHostController, searchViewModel: SearchViewModel) {
    @SuppressLint("SimpleDateFormat")
    val dateFormat = SimpleDateFormat("yyyy-MM-dd'T'HH:mm:ss.SSSSSSSSSZ")
    val scope = rememberCoroutineScope()
    var selectedOption by remember { mutableStateOf<IndexInfo?>(null) }
    val client = ApiClient(getApiKey(LocalContext.current))
    var isLoading by remember { mutableStateOf(false) }

    Scaffold(
        topBar = {
            CenterAlignedTopAppBar(
                title = {
                    Text("Index")
                },
                navigationIcon = {
                    IconButton(onClick = {
                        navHostController.navigate("settings")
                    }) {
                        Icon(
                            imageVector = Icons.Filled.ArrowBack,
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
            if (searchViewModel.currentChannel == null) {
                Text(
                    text = "No channel selected, go back to the previous screen",
                    style = MaterialTheme.typography.labelMedium,
                    modifier = Modifier.padding(bottom = 16.dp)
                )
            } else {
                if (searchViewModel.currentChannel!!.indices.isEmpty()) {
                    Column(
                        modifier = Modifier
                            .fillMaxSize()
                            .padding(16.dp),
                        horizontalAlignment = Alignment.CenterHorizontally
                    ) {
                        Text(
                            text = "No Indices for this channel",
                            fontSize = 30.sp,
                            fontWeight = FontWeight.Bold,
                            color = Color.Black
                        )
                        Text(
                            text = "Generation of an index can take up to 10 minutes, are you ready?",
                            fontSize = 18.sp,
                            color = Color.Gray,
                            modifier = Modifier.padding(top = 8.dp, bottom = 32.dp)
                        )

                        if (isLoading) {
                            CircularProgressIndicator(
                                modifier = Modifier
                                    .fillMaxSize()
                                    .padding(16.dp),
                            )
                        } else {
                            Button(
                                onClick = {
                                    scope.launch {
                                        Log.d(
                                            "index",
                                            "Generating index for ${searchViewModel.currentChannel!!.name}"
                                        )

                                        try {
                                            client.generateIndex(searchViewModel.currentChannel!!.name)
                                        } catch (e: Exception) {
                                            Log.d("index", "Failed to index $e")
                                            cancel()
                                        }

                                        Log.d("index", "Generating done")
                                        searchViewModel.populateData(client)
                                        navHostController.navigate("index")
                                    }
                                },
                                shape = RectangleShape,
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .padding(16.dp)
                            ) {
                                Text("Generate Index", fontSize = 18.sp)
                            }
                        }
                    }
                } else {
                    Text(
                        text = "${searchViewModel.currentChannel!!.indices.size} AVAILABLE INDICES FOR ${searchViewModel.currentChannel!!.name}",
                        style = MaterialTheme.typography.labelMedium,
                        modifier = Modifier.padding(bottom = 16.dp)
                    )
                    LazyColumn {
                        itemsIndexed(searchViewModel.currentChannel!!.indices) { index, item ->
                            Row(
                                verticalAlignment = Alignment.CenterVertically,
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .padding(vertical = 8.dp)
                            ) {
                                RadioButton(
                                    selected = selectedOption == item, onClick = {
                                        searchViewModel.currentIndex = item
                                        selectedOption = item
                                    }, modifier = Modifier.padding(end = 8.dp)
                                )
                                Text(item.date.toString())
                            }
                        }
                    }

                    if (isLoading) {
                        CircularProgressIndicator(
                            modifier = Modifier
                                .fillMaxSize()
                                .padding(16.dp),
                        )
                    } else {
                        Button(
                            onClick = {
                                scope.launch {
                                    Log.d(
                                        "index",
                                        "Generating index for ${searchViewModel.currentChannel!!.name}"
                                    )

                                    isLoading = true
                                    try {
                                        client.generateIndex(searchViewModel.currentChannel!!.name)
                                    } catch (e: Exception) {
                                        Log.d("index", "Failed to index $e")
                                        cancel()
                                    }
                                    isLoading = false

                                    Log.d("index", "Generating done")
                                    searchViewModel.populateData(client)
                                    navHostController.navigate("index")
                                }
                            },
                            shape = RectangleShape,
                            modifier = Modifier
                                .fillMaxWidth()
                                .padding(16.dp)
                        ) {
                            Text("Generate Index", fontSize = 18.sp)
                        }
                    }
                }
            }
        }
    }
}