package com.example.nixhund

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.runtime.Composable
import androidx.compose.ui.platform.LocalContext
import androidx.navigation.NavHostController
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import com.example.nixhund.screens.History
import com.example.nixhund.screens.Login
import com.example.nixhund.screens.Search
import com.example.nixhund.screens.Settings


class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContent {
            val navController = rememberNavController()
            Navigation(navController)
        }
    }
}

@Composable
fun Navigation(navController: NavHostController) {
    setLoggedIn(LocalContext.current, false)

    val startDestination = if (getLoggedIn(LocalContext.current)) "search" else "welcome"
    NavHost(navController = navController, startDestination = startDestination) {
        composable(route = "search") { Search(navController) }
        composable(route = "settings") { Settings(navController) }
        composable(route = "channel") { Settings(navController) }
        composable(route = "index") { Settings(navController) }
        composable(route = "history") { History(navController) }
        composable(route = "welcome") { Settings(navController) }
        composable(route = "register") { Settings(navController) }
        composable(route = "login") { Login(navController) }
    }
}