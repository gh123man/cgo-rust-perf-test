extern crate alloc;
extern crate wee_alloc;
use ::value::{Secrets, Value};
use lazy_static::lazy_static;
use regex::Regex;
use std::cell::RefCell;
use std::collections::BTreeMap;
use std::ffi::CStr;
use std::ffi::CString;
use vrl::Program;
use vrl::TimeZone;
use vrl::{state, Runtime, TargetValueRef};

use alloc::vec::Vec;
use std::mem::MaybeUninit;
use std::slice;

lazy_static! {
    static ref RE: Regex = Regex::new(r"\b\w{4}\b").unwrap();
    static ref VRL_PROGRAM: Program = compile_vrl();
}

thread_local! {static RUNTIME: RefCell<Runtime> = RefCell::new(Runtime::new(state::Runtime::default()));}

pub fn compile_vrl() -> Program {
    // let program = r#"."#;
    let program = r#". = replace(string!(.), r'\b\w{4}\b', "rust", 1)"#;
    return vrl::compile(&program, &vrl_stdlib::all()).unwrap().program;
}

pub fn run_vrl(s: &str) -> String {
    let mut value: Value = Value::from(s);
    let mut metadata = Value::Object(BTreeMap::new());
    let mut secrets = Secrets::new();
    let mut target = TargetValueRef {
        value: &mut value,
        metadata: &mut metadata,
        secrets: &mut secrets,
    };

    let output = RUNTIME.with(|r| {
        // r.borrow_mut().clear();
        return r
            .borrow_mut()
            .resolve(&mut target, &VRL_PROGRAM, &TimeZone::Local);
    });

    return output.unwrap().to_string();
}

#[no_mangle]
pub extern "C" fn transform(input: *const libc::c_char) -> *const libc::c_char {
    let inpt: &CStr = unsafe { CStr::from_ptr(input) };
    let replaced = RE.replacen(inpt.to_str().unwrap(), 1, "rust");
    let c_str = CString::new(replaced.as_bytes()).expect("CString::new failed");
    return c_str.into_raw();
}

#[no_mangle]
pub extern "C" fn noop(input: *const libc::c_char) -> *const libc::c_char {
    let inpt: &CStr = unsafe { CStr::from_ptr(input) };
    let c_str = CString::new(inpt.to_str().unwrap()).expect("CString::new failed");
    return c_str.into_raw();
}

#[no_mangle]
pub extern "C" fn transform_vrl(input: *const libc::c_char) -> *const libc::c_char {
    let inpt: &CStr = unsafe { CStr::from_ptr(input) };
    let output = run_vrl(inpt.to_str().unwrap());
    let c_str = CString::new(output.as_bytes()).expect("CString::new failed");
    return c_str.into_raw();
}

// Wasm Integration Below
/// WebAssembly export that accepts a string (linear memory offset, byteCount)
/// and returns a pointer/size pair packed into a u64.
///
/// Note: The return value is leaked to the caller, so it must call
/// [`deallocate`] when finished.
/// Note: This uses a u64 instead of two result values for compatibility with
/// WebAssembly 1.0.
#[cfg_attr(all(target_arch = "wasm32"), export_name = "noop_wasm")]
#[no_mangle]
pub unsafe extern "C" fn _noop_wasm(ptr: u32, len: u32) -> u64 {
    let name = &ptr_to_string(ptr, len);
    let new_string = String::from(name); // the no-op
    let (ptr, len) = string_to_ptr(&new_string);
    // Note: This changes ownership of the pointer to the external caller. If
    // we didn't call forget, the caller would read back a corrupt value. Since
    // we call forget, the caller must deallocate externally to prevent leaks.
    std::mem::forget(new_string);
    return ((ptr as u64) << 32) | len as u64;
}

// WASM String-related helper functions
/// Returns a string from WebAssembly compatible numeric types representing
/// its pointer and length.
unsafe fn ptr_to_string(ptr: u32, len: u32) -> String {
    let slice = slice::from_raw_parts_mut(ptr as *mut u8, len as usize);
    let utf8 = std::str::from_utf8_unchecked_mut(slice);
    return String::from(utf8);
}

/// Returns a pointer and size pair for the given string in a way compatible
/// with WebAssembly numeric types.
///
/// Note: This doesn't change the ownership of the String. To intentionally
/// leak it, use [`std::mem::forget`] on the input after calling this.
unsafe fn string_to_ptr(s: &String) -> (u32, u32) {
    return (s.as_ptr() as u32, s.len() as u32);
}

// WASM Memory-related helper functinos
//
// TODO - Only enable WeeAlloc in wasm32 builds
// #[cfg_attr(all(target_arch = "wasm32"), global_allocator)]
/// Set the global allocator to the WebAssembly optimized one.
#[global_allocator]
static ALLOC: wee_alloc::WeeAlloc = wee_alloc::WeeAlloc::INIT;
/// WebAssembly export that allocates a pointer (linear memory offset) that can
/// be used for a string.
///
/// This is an ownership transfer, which means the caller must call
/// [`deallocate`] when finished.
#[cfg_attr(all(target_arch = "wasm32"), export_name = "allocate")]
#[no_mangle]
pub extern "C" fn _allocate(size: u32) -> *mut u8 {
    allocate(size as usize)
}

/// Allocates size bytes and leaks the pointer where they start.
fn allocate(size: usize) -> *mut u8 {
    // Allocate the amount of bytes needed.
    let vec: Vec<MaybeUninit<u8>> = Vec::with_capacity(size);

    // into_raw leaks the memory to the caller.
    Box::into_raw(vec.into_boxed_slice()) as *mut u8
}

/// WebAssembly export that deallocates a pointer of the given size (linear
/// memory offset, byteCount) allocated by [`allocate`].
#[cfg_attr(all(target_arch = "wasm32"), export_name = "deallocate")]
#[no_mangle]
pub unsafe extern "C" fn _deallocate(ptr: u32, size: u32) {
    deallocate(ptr as *mut u8, size as usize);
}

/// Retakes the pointer which allows its memory to be freed.
unsafe fn deallocate(ptr: *mut u8, size: usize) {
    // TODO - should this be Box::from_raw? (see Box::into_raw docs)
    let _ = Vec::from_raw_parts(ptr, 0, size);
}
