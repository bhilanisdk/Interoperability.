package car1

import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import rust.interop.bridge.FilterParams
import rust.interop.bridge.FilterResponse
import rust.interop.bridge.fetchInteroperability

val params = FilterParams(
    integration = null,
    developmentkit = null,
    language = null,
    crates = null,
    page = "1",
    ids = null
)

suspend fun fetchDataFromRust(pageNumber: Int): FilterResponse {
    params.page = pageNumber.toString();
    return withContext(Dispatchers.IO) {
        fetchInteroperability(params)
    }
}