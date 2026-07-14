package bhilani.interoperability.jvm;

public class JVMSDKit {
    static { System.loadLibrary("interoperability_wrapper_robusta"); }
    public native String fetchInteroperability(String url, String json);
}