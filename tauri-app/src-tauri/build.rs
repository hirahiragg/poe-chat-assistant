fn main() {
    tauri_build::build();

    #[cfg(target_os = "macos")]
    {
        cc::Build::new()
            .file("src/focus/macos.m")
            .flag("-fobjc-arc")
            .compile("focus_macos");
        println!("cargo:rustc-link-lib=framework=AppKit");
    }
}
