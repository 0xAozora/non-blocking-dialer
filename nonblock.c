#include <winsock2.h>
#include <stdint.h>
#include <stdbool.h>

bool set_nonblock(int fd, bool nonblocking) {
   if (fd < 0) return false;

   unsigned long mode = nonblocking ? 1 : 0;
   return (ioctlsocket(fd, FIONBIO, &mode) == 0);
}