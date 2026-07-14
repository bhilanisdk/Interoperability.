@Grab('org.codehaus.gpars:gpars:1.2.1')
import bhilani.interoperability.jvm.JVMSDKit
import groovy.json.*
import groovyx.gpars.GParsPool

// trick to print the warnings before the header
GParsPool.withPool(1) { [].collectParallel { it } }

def sdk = new JVMSDKit()
def rand = new Random()

println "--- Bhilani Interop SDK (Groovy Concurrency) ---"
def totalStart = System.currentTimeMillis()

GParsPool.withPool {
    def results = (1..5).collectParallel { pageNum ->
        def start = System.currentTimeMillis()
        try {
            sleep(rand.nextInt(201) + 50)
            
            def params = JsonOutput.toJson([page: "${pageNum}"])
            def response = sdk.fetchInteroperability("", params)
            def parsed = new JsonSlurper().parseText(response)
            
            return [
                pageNum: pageNum, 
                success: true, 
                data: parsed.data, 
                total: parsed.pagination.total_pages,
                duration: System.currentTimeMillis() - start
            ]
        } catch (e) {
            return [
                pageNum: pageNum, 
                success: false, 
                error: e.message,
                duration: System.currentTimeMillis() - start
            ]
        }
    }

    results.each { res ->
        if (res.success) {
            if (!res.data || res.pageNum > res.total) {
                println "Page ${res.pageNum}: Success (No Data) [${res.duration}ms]"
            } else {
                println "Page ${res.pageNum}: Success [${res.duration}ms]"
                res.data.each { println "  - Title: ${it.title}" }
            }
        } else {
            println "Page ${res.pageNum}: Failed (${res.error}) [${res.duration}ms]"
        }
    }
}

println "\nTotal session duration: ${System.currentTimeMillis() - totalStart}ms"