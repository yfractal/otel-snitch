require 'mkmf'
$CFLAGS << ' -Wall -Wextra -Wno-unused-parameter -Wno-missing-field-initializers'
create_makefile("otel_snitch/otel_snitch")
