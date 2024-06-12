package com.example.nixhund.screens

import android.content.ClipData
import android.content.ClipboardManager
import android.content.Context
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material3.CenterAlignedTopAppBar
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.SnackbarHost
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.material3.rememberTopAppBarState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.unit.dp
import androidx.navigation.NavHostController
import com.example.nixhund.R
import com.example.nixhund.SearchViewModel
import com.example.nixhund.api.PkgResult
import com.example.nixhund.ui.theme.Purple80
import kotlinx.coroutines.cancel
import kotlinx.coroutines.launch

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun Detail(navHostController: NavHostController, searchViewModel: SearchViewModel) {
    val scrollBehavior = TopAppBarDefaults.pinnedScrollBehavior(rememberTopAppBarState())
    val pkg: PkgResult = searchViewModel.currentPackage!!
    val snackbarHostState = remember { SnackbarHostState() }

    Scaffold(topBar = {
        CenterAlignedTopAppBar(
            colors = TopAppBarDefaults.topAppBarColors(),
            title = {
                Text(text = "Package details")
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
            scrollBehavior = scrollBehavior,
        )
    }, snackbarHost = {
        SnackbarHost(hostState = snackbarHostState)
    }) { contentPadding ->
        Column(modifier = Modifier.padding(contentPadding)) {
            Column(
                modifier = Modifier.padding(16.dp),
                horizontalAlignment = Alignment.CenterHorizontally
            ) {
                DetailSection(title = "Package Name", content = pkg.pkgName, snackbarHostState)
                Spacer(modifier = Modifier.height(16.dp))
                DetailSection(
                    title = "NixOS Configuration",
                    content = "environment.systemPackages = [ ${pkg.pkgName} ];",
                    snackbarHostState
                )
                Spacer(modifier = Modifier.height(16.dp))
                DetailSection(
                    title = "nix-shell",
                    content = "nix-shell -p ${pkg.pkgName}",
                    snackbarHostState
                )
                Spacer(modifier = Modifier.height(16.dp))
                DetailInfo(title = "Version", content = "v1.0")
                Spacer(modifier = Modifier.height(8.dp))
                DetailInfo(title = "File", content = pkg.path)
                Spacer(modifier = Modifier.height(8.dp))
                DetailInfo(title = "Output hash", content = pkg.outHash)
                Spacer(modifier = Modifier.height(8.dp))
                DetailInfo(title = "Output name", content = pkg.outName)
            }
        }
    }
}

@Composable
fun DetailSection(title: String, content: String, snackbarHostState: SnackbarHostState) {
    val scope = rememberCoroutineScope()
    val context = LocalContext.current
    Column {
        Text(text = title, style = MaterialTheme.typography.labelSmall)
        Spacer(modifier = Modifier.height(4.dp))
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .background(Purple80, shape = RoundedCornerShape(8.dp))
                .padding(8.dp)
        ) {
            Row(
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.SpaceBetween,
                modifier = Modifier.fillMaxWidth()
            ) {
                Text(text = content, style = MaterialTheme.typography.labelSmall, maxLines = 1)
                IconButton(onClick = {
                    scope.launch {
                        try {
                            copyToClipboard(context, content)
                        } catch (e: Exception) {
                            snackbarHostState.showSnackbar("Copying to clipboard failed miserably")
                            cancel()
                        }
                        snackbarHostState.showSnackbar("Copied to clipboard")
                    }
                }) {
                    Icon(
                        painter = painterResource(id = R.drawable.copy_to_clipboard),
                        "Copy to clipboard"
                    )
                }
            }
        }
    }
}

@Composable
fun DetailInfo(title: String, content: String) {
    Column {
        Text(text = title, style = MaterialTheme.typography.labelMedium)
        Spacer(modifier = Modifier.height(4.dp))
        Text(
            text = content,
            style = MaterialTheme.typography.labelSmall,
            modifier = Modifier
                .fillMaxWidth()
                .background(Purple80, shape = RoundedCornerShape(8.dp))
                .padding(8.dp)
        )
    }
}

fun copyToClipboard(context: Context, text: CharSequence) {
    val clipboard = context.getSystemService(Context.CLIPBOARD_SERVICE) as ClipboardManager
    val clip = ClipData.newPlainText("label", text)
    clipboard.setPrimaryClip(clip)
}