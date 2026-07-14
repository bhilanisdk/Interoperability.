package bhilani.interoperability.jvm;

import java.util.*;
import java.util.concurrent.*;
import java.util.regex.*;
import java.util.stream.*;

public class JVMSDKit {
    private static boolean isLibLoaded = false;
    private static String platformInfo = "Unknown";

    static {
        String os = System.getProperty("os.name").toLowerCase();
        String arch = System.getProperty("os.arch").toLowerCase();
        platformInfo = os + " (" + arch + ")";

        try {
            boolean isSupportedOs = os.contains("win") || os.contains("mac") || os.contains("nux") || os.contains("nix");
            boolean isSupportedArch = arch.contains("64") || arch.contains("amd64") || arch.contains("aarch64");

            if (isSupportedOs && isSupportedArch) {
                System.loadLibrary("interoperability_wrapper_robusta");
                isLibLoaded = true;
            } else {
                System.err.println("Unsupported platform: " + platformInfo + ". Native features disabled.");
            }
        } catch (UnsatisfiedLinkError e) {
            System.err.println("Native library not found for " + platformInfo + ": " + e.getMessage());
        }
    }

    private native String fetchInteroperability(String url, String paramsJson);

    public boolean isReady() { return isLibLoaded; }

    public List<TimedResult> fetchPages(String url, int start, int end) {
        ExecutorService executor = Executors.newCachedThreadPool();
        Random random = new Random();

        List<CompletableFuture<TimedResult>> futures = IntStream.rangeClosed(start, end)
            .mapToObj(page -> CompletableFuture.supplyAsync(() -> {
                if (!isLibLoaded) return new TimedResult(page, null, new Exception("Library not loaded"), 0);

                long startTime = System.currentTimeMillis();
                try {
                    Thread.sleep(random.nextInt(201) + 50); // delay(50, 251)
                    String res = fetchInteroperability(url, String.format("{\"page\": \"%d\"}", page));
                    return new TimedResult(page, res, null, System.currentTimeMillis() - startTime);
                } catch (Exception e) {
                    return new TimedResult(page, null, e, System.currentTimeMillis() - startTime);
                }
            }, executor))
            .collect(Collectors.toList());

        return futures.stream().map(CompletableFuture::join).collect(Collectors.toList());
    }

    public static void main(String[] args) {
        JVMSDKit sdk = new JVMSDKit();
        long totalStart = System.currentTimeMillis();

        System.out.println("--- Bhilani Interop SDK (Java Concurrency) ---");

        if (!sdk.isReady()) {
            System.out.println("Abort: Native library not loaded for this platform.");
            return;
        }

        sdk.fetchPages("", 1, 5).forEach(res -> {
            if (res.error == null) {
                try {
                    int totalPages = parseTotalPages(res.data);
                    List<String> titles = parseTitles(res.data);

                    if (titles.isEmpty() || res.pageNum > totalPages) {
                        System.out.printf("Page %d: Success (No Data) [%dms]%n", res.pageNum, res.duration);
                    } else {
                        System.out.printf("Page %d: Success [%dms]%n", res.pageNum, res.duration);
                        titles.forEach(t -> System.out.println("  - Title: " + t));
                    }
                } catch (Exception e) {
                    System.out.printf("Page %d: Success (JSON Parsing Failed) [%dms]%n", res.pageNum, res.duration);
                }
            } else {
                System.out.printf("Page %d: Failed (%s) [%dms]%n", res.pageNum, res.error.getMessage(), res.duration);
            }
        });

        System.out.println("\nTotal session duration: " + (System.currentTimeMillis() - totalStart) + "ms");
    }

    private static int parseTotalPages(String json) {
        Matcher m = Pattern.compile("\"total_pages\"\\s*:\\s*(\\d+)").matcher(json);
        return m.find() ? Integer.parseInt(m.group(1)) : 0;
    }

    private static List<String> parseTitles(String json) {
        List<String> titles = new ArrayList<>();
        Matcher m = Pattern.compile("\"title\"\\s*:\\s*\"([^\"]+)\"").matcher(json);
        while (m.find()) titles.add(m.group(1));
        return titles;
    }

    static class TimedResult {
        int pageNum; String data; Exception error; long duration;
        TimedResult(int p, String d, Exception e, long dur) { 
            this.pageNum = p; this.data = d; this.error = e; this.duration = dur; 
        }
    }
}