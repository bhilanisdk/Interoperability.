package bhilani.interoperability.jvm

import kotlinx.coroutines.*
import kotlinx.serialization.*
import kotlinx.serialization.json.*
import kotlin.random.Random

@Serializable
data class Pagination(@SerialName("total_pages") val totalPages: Int)

@Serializable
data class SDKItem(val title: String)

@Serializable
data class FetchResponse(val data: List<SDKItem>, val pagination: Pagination)

data class TimedResult(val pageNum: Int, val result: Result<String>, val duration: Long)

class JVMSDKit {
    private external fun fetchInteroperability(url: String, paramsJson: String): String

    companion object {
        private var isLibLoaded = false
        private var platformInfo = "Unknown"

        init {
            val os = System.getProperty("os.name").lowercase()
            val arch = System.getProperty("os.arch").lowercase()
            platformInfo = "$os ($arch)"

            try {
                // OS Check: Windows, Mac, or Linux
                val isSupportedOs = os.contains("win") || os.contains("mac") || 
                                   os.contains("nix") || os.contains("nux")
                
                // Architecture Check: x64 or ARM64
                val isSupportedArch = arch.contains("64") || arch.contains("amd64") || 
                                     arch.contains("aarch64")

                if (isSupportedOs && isSupportedArch) {
                    System.loadLibrary("interoperability_wrapper_robusta")
                    isLibLoaded = true
                } else {
                    System.err.println("Unsupported platform: $platformInfo. Native features disabled.")
                }
            } catch (e: UnsatisfiedLinkError) {
                System.err.println("Native library not found for $platformInfo: ${e.message}")
            }
        }

        val jsonParser = Json { ignoreUnknownKeys = true }
    }

    fun isReady() = isLibLoaded

    suspend fun fetchPages(url: String, pageRange: IntRange): List<TimedResult> = coroutineScope {
        pageRange.map { page ->
            async(Dispatchers.IO) {
                if (!isLibLoaded) return@async TimedResult(page, Result.failure(Exception("Library not loaded")), 0)
                
                delay(Random.nextLong(50, 251))
                val start = System.currentTimeMillis()
                val res = runCatching { 
                    fetchInteroperability(url, """{"page": "$page"}""") 
                }
                val end = System.currentTimeMillis()
                TimedResult(page, res, end - start)
            }
        }.awaitAll()
    }
}

suspend fun main() {
    val sdk = JVMSDKit()
    val url = ""
    val totalStart = System.currentTimeMillis()

    println("--- Bhilani Interop SDK (Kotlin Concurrency) ---")
    
    if (!sdk.isReady()) {
        println("Abort: Native library not loaded for this platform.")
        return
    }

    sdk.fetchPages(url, 1..5).forEach { (pageNum, result, time) ->
        result.onSuccess { res ->
            try {
                val parsed = JVMSDKit.jsonParser.decodeFromString<FetchResponse>(res)
                val totalPages = parsed.pagination.totalPages

                if (parsed.data.isEmpty() || pageNum > totalPages) {
                    println("Page $pageNum: Success (No Data) [${time}ms]")
                } else {
                    println("Page $pageNum: Success [${time}ms]")
                    parsed.data.forEach { println("  - Title: ${it.title}") }
                }
            } catch (e: Exception) {
                println("Page $pageNum: Success (JSON Parsing Failed) [${time}ms]")
            }
        }
        result.onFailure { println("Page $pageNum: Failed (${it.message}) [${time}ms]") }
    }

    println("\nTotal session duration: ${System.currentTimeMillis() - totalStart}ms")
}