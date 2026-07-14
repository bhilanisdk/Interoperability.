//> using dep "com.lihaoyi::upickle:4.4.3"
//> using javaOpt "--enable-native-access=ALL-UNNAMED"
//> using javaProp "java.library.path=."

package bhilani.interoperability.jvm

import upickle.default.*
import upickle.implicits.key
import scala.concurrent.{Future, Await}
import scala.concurrent.duration.*
import scala.concurrent.ExecutionContext.Implicits.global
import scala.util.{Try, Success, Failure, Random}

case class Pagination(
  @key("total_pages") totalPages: Int
) derives ReadWriter

case class SDKItem(
  title: String
) derives ReadWriter

case class FetchResponse(
  data: List[SDKItem],
  pagination: Pagination
) derives ReadWriter

case class TimedResult(
  pageNum: Int, 
  result: Try[String], 
  duration: Long
)

class JVMSDKit:
  @native private def fetchInteroperability(url: String, paramsJson: String): String

  private val isLibLoaded: Boolean = 
    val os = System.getProperty("os.name").toLowerCase
    val arch = System.getProperty("os.arch").toLowerCase
    val isSupported = (os.contains("win") || os.contains("mac") || os.contains("nux")) && 
                      (arch.contains("64") || arch.contains("amd64") || arch.contains("aarch64"))
    
    if isSupported then
      Try(System.loadLibrary("interoperability_wrapper_robusta")).isSuccess
    else false

  def isReady(): Boolean = isLibLoaded

  def fetchPages(url: String, pageRange: Range): Future[List[TimedResult]] =
    val tasks = pageRange.toList.map { page =>
      Future:
        if !isLibLoaded then 
          TimedResult(page, Failure(Exception("Library not loaded")), 0)
        else
          Thread.sleep(Random.nextInt(201) + 50)
          val start = System.currentTimeMillis()
          val res = Try(fetchInteroperability(url, write(Map("page" -> page.toString))))
          TimedResult(page, res, System.currentTimeMillis() - start)
    }
    Future.sequence(tasks)

@main def runConcurrentSDK(): Unit =
  val sdk = JVMSDKit()
  val totalStart = System.currentTimeMillis()

  println("--- Bhilani Interop SDK (Scala Concurrency) ---")

  if !sdk.isReady() then
    println("Abort: Native library not loaded for this platform.")
  else
    val results = Await.result(sdk.fetchPages("", 1 to 5), 30.seconds)

    results.foreach { case TimedResult(pageNum, result, time) =>
      result match
        case Success(res) =>
          Try(read[FetchResponse](res)) match
            case Success(parsed) =>
              val totalPages = parsed.pagination.totalPages
              if parsed.data.isEmpty || pageNum > totalPages then
                println(s"Page $pageNum: Success (No Data) [${time}ms]")
              else
                println(s"Page $pageNum: Success [${time}ms]")
                parsed.data.foreach(item => println(s"  - Title: ${item.title}"))
            case Failure(_) =>
              println(s"Page $pageNum: Success (JSON Parsing Failed) [${time}ms]")
        case Failure(e) =>
          println(s"Page $pageNum: Failed (${e.getMessage}) [${time}ms]")
    }
    println(s"\nTotal session duration: ${System.currentTimeMillis() - totalStart}ms")