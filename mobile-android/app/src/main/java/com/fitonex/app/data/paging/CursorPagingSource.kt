package com.fitonex.app.data.paging

import androidx.paging.PagingSource
import androidx.paging.PagingState
import com.fitonex.app.data.model.PaginatedResponse

class CursorPagingSource<T : Any>(
    private val initialCursor: String? = null,
    private val fetch: suspend (cursor: String?) -> PaginatedResponse<T>
) : PagingSource<String, T>() {

    override suspend fun load(params: LoadParams<String>): LoadResult<String, T> {
        return try {
            val cursor = params.key ?: initialCursor
            val response = fetch(cursor)
            val nextCursor = response.next_cursor?.takeIf { response.has_more }
            LoadResult.Page(
                data = response.items,
                prevKey = null,
                nextKey = nextCursor
            )
        } catch (e: Exception) {
            LoadResult.Error(e)
        }
    }

    override fun getRefreshKey(state: PagingState<String, T>): String? {
        return state.anchorPosition?.let { position ->
            state.getClosestPageToPosition(position)?.prevKey
        }
    }
}
