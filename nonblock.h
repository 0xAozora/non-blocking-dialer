#ifndef NONBLOCK_H_
#define NONBLOCK_H_

#ifdef __cplusplus
extern "C" {
#endif

#include <stdint.h>
#include <stdbool.h>

// Windows-specific includes
#include <winsock2.h>
#include <ws2tcpip.h>

// Function declaration
bool set_nonblock(int fd, bool nonblocking);

#ifdef __cplusplus
} /* extern "C" */
#endif

#endif /* NONBLOCK_H_ */