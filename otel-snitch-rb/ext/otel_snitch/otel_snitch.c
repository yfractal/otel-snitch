#include <stdio.h>
#include <stdlib.h>
#include <fcntl.h>
#include <sys/mman.h>
#include <sys/stat.h>
#include <unistd.h>
#include <string.h>
#include <stdint.h>
#include <stdbool.h>
#include "ruby.h"
#include "data.h"
// Forward declarations

char *base_addr = NULL;

void* snitch_allocate(char **memory_addr, size_t size) {
    void *addr = *memory_addr;
    *memory_addr = (char *)addr + size;

    return addr;
}

int copy_rb_string_to_memory(VALUE key_str, char **memory_addr) {
    char *dest = (char *)(*memory_addr);
    strcpy(dest, RSTRING_PTR(key_str));
    *memory_addr = dest + strlen(dest) + 1;
    return (int)(dest - base_addr);
}

static int save_attribute(VALUE key, VALUE value, VALUE addr_ptr_val, KeyValue* attribute) {
    char **memory_addr_ptr = (char **)addr_ptr_val;

    if (TYPE(key) == T_SYMBOL) {
       attribute->key_offset = copy_rb_string_to_memory(rb_sym2str(key), memory_addr_ptr);
    } else {
       attribute->key_offset = copy_rb_string_to_memory(key, memory_addr_ptr);
    }

    size_t any_value_offset = (size_t)(*memory_addr_ptr - base_addr);
    attribute->value_offset = any_value_offset;
    AnyValue *any_value = snitch_allocate(memory_addr_ptr, sizeof(AnyValue));

    if (TYPE(value) == T_FIXNUM) {
        // any_value.value.int_value = NUM2INT(value);
        any_value->value_type = ANYVALUE_INT;
    } else if (TYPE(value) == T_STRING) {
        any_value->value_type = ANYVALUE_STRING;
        any_value->value.string_value_offset = copy_rb_string_to_memory(value, memory_addr_ptr);
    // } else if (TYPE(value) == T_FLOAT) {
    //     any_value.value.double_value = NUM2DBL(value);
    //     any_value.value_type = ANYVALUE_DOUBLE;
    // } else if (TYPE(value) == T_TRUE || TYPE(value) == T_FALSE) {
    //     any_value.value.bool_value = RTEST(value);
    //     any_value.value_type = ANYVALUE_BOOL;
    } else {
        rb_raise(rb_eTypeError, "Unsupported value type");
    }

    return 0;
}

int32_t save_hash(VALUE rb_hash, char **current_addr) {
    VALUE  rb_hash_arr = rb_funcall(rb_hash, rb_intern("to_a"), 0);
    int hash_pair_count = NUM2INT(rb_funcall(rb_hash_arr, rb_intern("count"), 0));

    KeyValue *key_vals = snitch_allocate(current_addr, hash_pair_count * sizeof(KeyValue));

    int j = 0;
    while (j < hash_pair_count) {
        VALUE pair = rb_ary_entry(rb_hash_arr, j);
        VALUE key = rb_ary_entry(pair, 0);
        VALUE value = rb_ary_entry(pair, 1);

        save_attribute(key, value, (VALUE)current_addr, &key_vals[j]);

        j++;
    }

    return (int32_t)((char *)key_vals - base_addr);
}

VALUE write_spans(VALUE self, VALUE file, VALUE attributes_file, VALUE spans) {
    const char *filepath = StringValueCStr(file);

    int fd = open(filepath, O_RDWR | O_CREAT, S_IRUSR | S_IWUSR);
    if (fd == -1) {
        perror("Failed to open file");
        return Qfalse;
    }
    long total = rb_array_len(spans);

    size_t data_len = MAX_FILE_SIZE;

    if (ftruncate(fd, data_len) == -1) {
        perror("Failed to set file size");
        close(fd);
        return Qfalse;
    }

    void *map = mmap(NULL, data_len, PROT_READ | PROT_WRITE, MAP_SHARED, fd, 0);
    base_addr = (char *)map;

    char *current_addr = map;

    if (map == MAP_FAILED) {
        perror("Failed to memory-map the file");
        close(fd);
        return Qfalse;
    }

    int64_t *span_offset_addr = snitch_allocate(&current_addr, MAX_SPANS * sizeof(int64_t));
    int spans_index = 0;

    int i = 0;
    while (i < total) {
        VALUE span = rb_ary_entry(spans, i);
        VALUE name = rb_iv_get(span, "@name");

        Span *span_ptr = (Span *)current_addr;
        current_addr = (char *)current_addr + sizeof(Span);

        int total_recorded_attributes = NUM2INT(rb_iv_get(span, "@total_recorded_attributes"));
        span_ptr->total_recorded_attributes = total_recorded_attributes;
        long str_offset = copy_rb_string_to_memory(name, &current_addr);
        span_ptr->name_offset = str_offset;

        VALUE kind = rb_funcall(self, rb_intern("span_kind_to_int"), 1, span);
        span_ptr->kind = NUM2INT(kind);

        VALUE status_code = rb_funcall(self, rb_intern("status_code"), 1, span);
        VALUE status_description = rb_funcall(self, rb_intern("status_description"), 1, span);

        Status *status = snitch_allocate(&current_addr, sizeof(Status));
        status->code = NUM2INT(status_code);
        status->description_offset = copy_rb_string_to_memory(status_description, &current_addr);
        span_ptr->status_offset = (int32_t)((char *)status - base_addr);

        // parent_span_id
        VALUE parent_span_id = rb_iv_get(span, "@parent_span_id");
        span_ptr->parent_span_id_offset = copy_rb_string_to_memory(parent_span_id, &current_addr);

        // start_timestamp
        VALUE start_timestamp = rb_iv_get(span, "@start_timestamp");
        span_ptr->start_timestamp_offset = copy_rb_string_to_memory(rb_big2str(start_timestamp, 10), &current_addr);

        // end_timestamp
        VALUE end_timestamp = rb_iv_get(span, "@end_timestamp");
        span_ptr->end_timestamp_offset = copy_rb_string_to_memory(rb_big2str(end_timestamp, 10), &current_addr);

        // attributes
        span_ptr->attributes_offset = save_hash(rb_iv_get(span, "@attributes"), &current_addr);

        // resource
        span_ptr->resource.attributes_offset = save_hash(rb_funcall(self, rb_intern("resource_attributes"), 1, span), &current_addr);

        // instrumentation_scope
        VALUE instrumentation_scope = rb_iv_get(span, "@instrumentation_scope");
        span_ptr->instrumentation_scope_offset = save_hash(instrumentation_scope, &current_addr);

        // span_id
        VALUE span_id = rb_funcall(self, rb_intern("span_id"), 1, span);
        span_ptr->span_id_offset = copy_rb_string_to_memory(span_id, &current_addr);

        // trace_id
        VALUE trace_id = rb_funcall(self, rb_intern("trace_id"), 1, span);
        span_ptr->trace_id_offset = copy_rb_string_to_memory(trace_id, &current_addr);

        // tracestate
        VALUE tracestate = rb_funcall(self, rb_intern("tracestate_str"), 1, span);
        span_ptr->tracestate_offset = copy_rb_string_to_memory(tracestate, &current_addr);

        // trace_flags
        VALUE trace_flags = rb_funcall(self, rb_intern("trace_flags"), 1, span);
        span_ptr->trace_flags = NUM2INT(trace_flags);

        // printf("[debug] span_offset is %lld\n", (int64_t)((char *)span_ptr - base_addr));
        span_offset_addr[spans_index++] = (int64_t)((char *)span_ptr - base_addr);

        i++;
    }

    return Qtrue;
}

// Initialize the extension
void Init_otel_snitch() {
    VALUE mOtelSnitch = rb_define_module("OtelSnitch");
    VALUE cExporter = rb_define_class_under(mOtelSnitch, "Exporter", rb_cObject);
    rb_define_method(cExporter, "write_spans", write_spans, 3);

}
