using System.Diagnostics;
using System.Runtime.InteropServices;
using System.Text.Json;
using System.Text.Json.Serialization;
using SDK = InteroperabilityWrapperRnet.InteroperabilityWrapperRnet;

namespace Bhilani.Interoperability;

public record SDKItem(
    [property: JsonPropertyName("title")] string Title
);

public record Pagination(
    [property: JsonPropertyName("total_pages")] int TotalPages
);

public record FetchResponse(
    [property: JsonPropertyName("data")] List<SDKItem> Data,
    [property: JsonPropertyName("pagination")] Pagination Pagination
);

public record Result<T>(T? Value, Exception? Error)
{
    public bool IsSuccess => Error == null;
    public static Result<T> Success(T value) => new(value, null);
    public static Result<T> Failure(Exception e) => new(default, e);
}

public record TimedResult(int PageNum, Result<string> Result, long Duration);

public class DotnetSDKit
{
    private static readonly bool IsLibLoaded;

    static DotnetSDKit()
    {
        try
        {
            var arch = RuntimeInformation.ProcessArchitecture;
            bool isSupportedOs = RuntimeInformation.IsOSPlatform(OSPlatform.Windows) ||
                                 RuntimeInformation.IsOSPlatform(OSPlatform.OSX) ||
                                 RuntimeInformation.IsOSPlatform(OSPlatform.Linux);

            bool isSupportedArch = arch == Architecture.X64 || arch == Architecture.Arm64;

            if (isSupportedOs && isSupportedArch)
            {
                IsLibLoaded = true;
            }
            else
            {
                Console.Error.WriteLine("Unsupported platform. Native features disabled.");
            }
        }
        catch (Exception e)
        {
            Console.Error.WriteLine($"Native library check failed: {e.Message}");
        }
    }

    public bool IsReady() => IsLibLoaded;

    public async Task<List<TimedResult>> FetchPagesAsync(IEnumerable<int> pageRange)
    {
        var tasks = pageRange.Select(async page =>
        {
            if (!IsLibLoaded) 
                return new TimedResult(page, Result<string>.Failure(new Exception("Library not loaded")), 0);

            await Task.Delay(Random.Shared.Next(50, 251));
            var sw = Stopwatch.StartNew();

            try
            {
                var paramsJson = $"{{\"page\": \"{page}\"}}";
                var response = await Task.Run(() => SDK.FetchForDotnet(paramsJson));
                return new TimedResult(page, Result<string>.Success(response), sw.ElapsedMilliseconds);
            }
            catch (Exception ex)
            {
                return new TimedResult(page, Result<string>.Failure(ex), sw.ElapsedMilliseconds);
            }
        });

        return (await Task.WhenAll(tasks)).ToList();
    }
}

class Program
{
    static async Task Main()
    {
        var sdk = new DotnetSDKit();
        var jsonOptions = new JsonSerializerOptions { PropertyNameCaseInsensitive = true };
        var totalSw = Stopwatch.StartNew();

        Console.WriteLine("--- Bhilani Interop SDK (.NET Concurrency) ---\n");

        if (!sdk.IsReady())
        {
            Console.WriteLine("Abort: Native library not loaded for this platform.");
            return;
        }

        var results = await sdk.FetchPagesAsync(Enumerable.Range(1, 5));

        foreach (var (pageNum, result, time) in results)
        {
            if (result.IsSuccess && result.Value != null)
            {
                try
                {
                    var parsed = JsonSerializer.Deserialize<FetchResponse>(result.Value, jsonOptions);
                    int totalPages = parsed?.Pagination.TotalPages ?? 0;

                    if (parsed?.Data == null || parsed.Data.Count == 0 || pageNum > totalPages)
                    {
                        Console.WriteLine($"Page {pageNum,2}: Success (No Data) [{time}ms]");
                    }
                    else
                    {
                        Console.WriteLine($"Page {pageNum,2}: Success [{time}ms]");
                        foreach (var item in parsed.Data) 
                            Console.WriteLine($"      - Title: {item.Title}");
                    }
                }
                catch
                {
                    Console.WriteLine($"Page {pageNum,2}: Success (JSON Parsing Failed) [{time}ms]");
                }
            }
            else
            {
                Console.WriteLine($"Page {pageNum,2}: Failed ({result.Error?.Message}) [{time}ms]");
            }
        }

        Console.WriteLine($"\nTotal session duration: {totalSw.ElapsedMilliseconds}ms");
    }
}