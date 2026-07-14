package rust.android.rustjni

import android.os.Bundle
import android.widget.Toast
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.text.selection.SelectionContainer
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalClipboardManager
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.AnnotatedString
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.google.gson.GsonBuilder
import com.google.gson.JsonObject
import rust.interop.bridge.FilterResponse
import rust.interop.bridge.Interoperability

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContent {
            var currentPage by remember { mutableIntStateOf(1) }
            var resultState by remember { mutableStateOf<FilterResponse?>(null) }
            var isLoading by remember { mutableStateOf(false) }
            var showRawJson by remember { mutableStateOf(false) }

            val totalPages = resultState?.pagination?.totalPages?.toInt() ?: 1
            val context = LocalContext.current

            LaunchedEffect(currentPage) {
                isLoading = true
                try {
                    resultState = fetchDataFromRust(currentPage)
                } catch (e: Exception) {
                    Toast.makeText(context, "Error: ${e.message}", Toast.LENGTH_SHORT).show()
                } finally {
                    isLoading = false
                }
            }

            MaterialTheme {
                Surface(modifier = Modifier.fillMaxSize()) {
                    Column(modifier = Modifier.padding(16.dp)) {
                        HeaderSection()

                        PaginationControls(
                            currentPage = currentPage,
                            totalPages = totalPages,
                            isLoading = isLoading,
                            onPageChange = { currentPage = it }
                        )

                        if (isLoading) {
                            LinearProgressIndicator(modifier = Modifier.fillMaxWidth().padding(bottom = 8.dp))
                        }

                        // Data List Section
                        Box(modifier = Modifier.weight(1f)) {
                            resultState?.let { ResultList(it.data) }
                        }

                        // Debug/JSON Section
                        JsonViewerSection(
                            resultState = resultState,
                            isVisible = showRawJson,
                            onToggle = { showRawJson = !showRawJson }
                        )
                    }
                }
            }
        }
    }
}

/** 1. Header Design **/
@Composable
fun HeaderSection() {
    Column {
        Text("BHILANI Interop SDK", style = MaterialTheme.typography.headlineMedium)
        Text("RUST, ANDROID NDK", style = MaterialTheme.typography.headlineMedium, color = Color.Gray)
    }
}

/** 2. Pagination Buttons Design **/
@Composable
fun PaginationControls(currentPage: Int, totalPages: Int, isLoading: Boolean, onPageChange: (Int) -> Unit) {
    Row(
        modifier = Modifier.fillMaxWidth().padding(vertical = 12.dp),
        horizontalArrangement = Arrangement.SpaceBetween,
        verticalAlignment = Alignment.CenterVertically
    ) {
        Button(
            onClick = { onPageChange(currentPage - 1) },
            enabled = !isLoading && currentPage > 1
        ) { Text("Previous") }

        Text("Page $currentPage of $totalPages", style = MaterialTheme.typography.bodyLarge)

        Button(
            onClick = { onPageChange(currentPage + 1) },
            enabled = !isLoading && currentPage < totalPages
        ) { Text("Next") }
    }
}

/** 3. Data Card Design **/
@Composable
fun ResultList(items: List<Interoperability>) {
    LazyColumn(modifier = Modifier.fillMaxSize()) {
        items(items) { item ->
            Card(
                modifier = Modifier.fillMaxWidth().padding(vertical = 4.dp),
                colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceVariant)
            ) {
                Column(modifier = Modifier.padding(12.dp)) {
                    Text(item.title, style = MaterialTheme.typography.titleMedium)
                    Text("Integration: ${item.integration}", style = MaterialTheme.typography.bodyMedium)
                }
            }
        }
    }
}

/** 4. Formatted JSON Viewer Design **/
@Composable
fun JsonViewerSection(resultState: FilterResponse?, isVisible: Boolean, onToggle: () -> Unit) {
    val clipboardManager = LocalClipboardManager.current
    val context = LocalContext.current
    val prettyGson = remember { GsonBuilder().setPrettyPrinting().create() }

    Column {
        Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.End) {
            if (isVisible && resultState != null) {
                TextButton(onClick = {
                    val json = prettyGson.toJson(resultState)
                    clipboardManager.setText(AnnotatedString(json))
                    Toast.makeText(context, "Copied to clipboard", Toast.LENGTH_SHORT).show()
                }) { Text("Copy") }
            }
            TextButton(onClick = onToggle) {
                Text(if (isVisible) "Hide Raw JSON" else "View Raw JSON")
            }
        }

        if (isVisible) {
            SelectionContainer {
                Surface(
                    modifier = Modifier.fillMaxWidth().heightIn(max = 250.dp),
                    color = Color(0xFFF4F4F4),
                    shape = MaterialTheme.shapes.small,
                    border = CardDefaults.outlinedCardBorder()
                ) {
                    Column(modifier = Modifier.padding(12.dp).verticalScroll(rememberScrollState())) {
                        Text(
                            text = resultState?.let { response ->
                                val orderedJson = JsonObject().apply {
                                    addProperty("message", response.message)
                                    add("pagination", prettyGson.toJsonTree(response.pagination))
                                    add("data", prettyGson.toJsonTree(response.data))
                                }
                                prettyGson.toJson(orderedJson)
                            } ?: "No data available",
                            fontFamily = FontFamily.Monospace,
                            fontSize = 11.sp,
                            color = Color(0xFF333333)
                        )
                    }
                }
            }
        }
    }
}