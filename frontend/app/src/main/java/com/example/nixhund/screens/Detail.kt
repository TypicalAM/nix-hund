package com.example.nixhund.screens

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
import androidx.compose.material.icons.filled.AddCircle
import androidx.compose.material3.CenterAlignedTopAppBar
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.material3.rememberTopAppBarState
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp
import androidx.navigation.NavHostController
import com.example.nixhund.SearchViewModel
import com.example.nixhund.api.PkgResult

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun Detail(navHostController: NavHostController, searchViewModel: SearchViewModel) {
    val scrollBehavior = TopAppBarDefaults.pinnedScrollBehavior(rememberTopAppBarState())
    val pkg: PkgResult = searchViewModel.currentPackage!!

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
    }) { contentPadding ->
        Column(
            modifier = Modifier.padding(contentPadding),
            verticalArrangement = Arrangement.Top,
            horizontalAlignment = Alignment.Start
        ) {
            DetailSection(title = "Package Name", content = pkg.pkgName)
            Spacer(modifier = Modifier.height(16.dp))
            DetailSection(
                title = "NixOS Configuration",
                content = "environment.systemPackages = [ ${pkg.pkgName} ];"
            )
            Spacer(modifier = Modifier.height(16.dp))
            DetailSection(title = "nix-shell", content = "nix-shell -p ${pkg.pkgName}")
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

@Composable
fun DetailSection(title: String, content: String) {
    Column {
        Text(text = title, style = MaterialTheme.typography.labelSmall)
        Spacer(modifier = Modifier.height(4.dp))
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .background(Color.LightGray, shape = RoundedCornerShape(8.dp))
                .padding(8.dp)
        ) {
            Row(
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.SpaceBetween,
                modifier = Modifier.fillMaxWidth()
            ) {
                Text(text = content, style = MaterialTheme.typography.labelSmall, maxLines = 1)
                IconButton(onClick = { /* handle copy */ }) {
                    Icon(Icons.Default.AddCircle, contentDescription = "Copy")
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
                .background(Color.LightGray, shape = RoundedCornerShape(8.dp))
                .padding(8.dp)
        )
    }
}