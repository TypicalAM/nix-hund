package com.example.nixhund

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.viewModels
import androidx.compose.runtime.Composable
import androidx.compose.ui.platform.LocalContext
import androidx.navigation.NavHostController
import androidx.navigation.NavType
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import androidx.navigation.navArgument
import com.example.nixhund.screens.Channel
import com.example.nixhund.screens.Detail
import com.example.nixhund.screens.History
import com.example.nixhund.screens.Index
import com.example.nixhund.screens.Login
import com.example.nixhund.screens.Register
import com.example.nixhund.screens.Search
import com.example.nixhund.screens.Settings
import com.example.nixhund.screens.Welcome


class MainActivity : ComponentActivity() {
    private val searchViewModel: SearchViewModel by viewModels()

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContent {
            val navController = rememberNavController()
            Navigation(navController, searchViewModel)
        }
    }
}

@Composable
fun Navigation(navController: NavHostController, searchViewModel: SearchViewModel) {
    val startDestination = if (getLoggedIn(LocalContext.current)) "search" else "welcome"
    NavHost(navController = navController, startDestination = startDestination) {
        composable(route = "search") { Search(navController, searchViewModel) }
        composable(route = "settings") { Settings(navController) }
        composable(route = "channel") { Channel(navController, searchViewModel) }
        composable(route = "index") { Index(navController, searchViewModel) }
        composable(route = "history") { History(navController) }
        composable(route = "welcome") { Welcome(navController) }
        composable(route = "register") { Register(navController, searchViewModel) }
        composable(route = "login") { Login(navController, searchViewModel) }
        composable(route = "detail") { Detail(navController, searchViewModel) }
    }
}