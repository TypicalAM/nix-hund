package com.example.nixhund.screens

import android.util.Log
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
import androidx.compose.material.icons.filled.Settings
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.DrawerValue
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.FloatingActionButton
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.ModalDrawerSheet
import androidx.compose.material3.ModalNavigationDrawer
import androidx.compose.material3.Scaffold
import androidx.compose.material3.SnackbarHost
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.material3.rememberDrawerState
import androidx.compose.material3.rememberTopAppBarState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.input.nestedscroll.nestedScroll
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.navigation.NavHostController
import com.example.nixhund.R
import com.example.nixhund.SearchViewModel
import com.example.nixhund.api.ApiClient
import com.example.nixhund.api.PkgResult
import com.example.nixhund.getApiKey
import com.example.nixhund.ui.Sidebar
import com.example.nixhund.ui.TopBarWithSearch
import kotlinx.coroutines.cancel
import kotlinx.coroutines.launch

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun Search(navHostController: NavHostController, searchViewModel: SearchViewModel) {
    val drawerState = rememberDrawerState(initialValue = DrawerValue.Closed)
    val scope = rememberCoroutineScope()
    val scrollBehavior = TopAppBarDefaults.pinnedScrollBehavior(rememberTopAppBarState())
    val snackbarHostState = remember { SnackbarHostState() }
    val client = ApiClient(getApiKey(LocalContext.current))

    var shownPkgs by remember { mutableStateOf<List<PkgResult>>(listOf()) }
    var isSearching by remember { mutableStateOf(false) }
    var isLoading by remember { mutableStateOf(false) }

    if (!searchViewModel.populated) searchViewModel.populateData(client)

    ModalNavigationDrawer(
        drawerState = drawerState,
        drawerContent = {
            ModalDrawerSheet {
                Sidebar(navHostController)
            }
        },
    ) {
        Scaffold(modifier = Modifier.nestedScroll(scrollBehavior.nestedScrollConnection), topBar = {
            TopBarWithSearch(navOnClick = {
                scope.launch { drawerState.apply { if (isClosed) open() else close() } }
            }, onSearch = { query ->
                scope.launch {
                    if (searchViewModel.currentIndex == null) {
                        snackbarHostState.showSnackbar("No index selected! Go to to the settings")
                        cancel()
                    }

                    isSearching = true
                    isLoading = true
                    val uuid = searchViewModel.currentIndex!!.id
                    var pkgs: List<PkgResult> = listOf()
                    try {
                        pkgs = client.indexQuery(uuid, query)
                        isLoading = false
                    } catch (e: Exception) {
                        snackbarHostState.showSnackbar("An error occurred when fetching the packages")
                        cancel()
                    }

                    shownPkgs = pkgs
                    isLoading = false
                    Log.d("search", "Found ${pkgs.size} results")
                }
            })
        }, snackbarHost = {
            SnackbarHost(hostState = snackbarHostState)
        }, floatingActionButton = {
            FloatingActionButton(onClick = { navHostController.navigate("settings") }) {
                Icon(Icons.Default.Settings, contentDescription = "Settings")
            }
        }) { contentPadding ->
            Column(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(contentPadding),
                verticalArrangement = Arrangement.Top,
                horizontalAlignment = Alignment.CenterHorizontally
            ) {
                if (!isSearching) {
                    if (searchViewModel.currentIndex == null) {
                        Text(
                            "Go into the settings and select an index, then come here and press the search icon to begin searching",
                            fontSize = 18.sp,
                            modifier = Modifier.padding(16.dp)
                        )
                    } else {
                        Text(
                            "Press the search icon to begin searching on ${searchViewModel.currentIndex!!.id}. This index was built at ${searchViewModel.currentIndex!!.date}",
                            fontSize = 12.sp,
                            modifier = Modifier.padding(16.dp)
                        )
                    }
                } else {
                    if (isLoading) {
                        CircularProgressIndicator(
                            modifier = Modifier.padding(16.dp),
                        )
                    } else {
                        LazyColumn {
                            itemsIndexed(shownPkgs) { _, item ->
                                PackageListItem(item = item, onClick = {
                                    searchViewModel.currentPackage = item
                                    navHostController.navigate("detail")
                                })
                            }
                        }
                    }
                }
            }
        }
    }
}

@Composable
fun PackageListItem(item: PkgResult, onClick: () -> Unit) {
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
            Text(text = item.pkgName, style = MaterialTheme.typography.labelSmall)
            Text(text = item.version, style = MaterialTheme.typography.labelSmall)
            Text(
                text = item.path,
                style = MaterialTheme.typography.labelSmall,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis
            )
        }
    }
}