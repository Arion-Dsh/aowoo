#ifndef _AOWOO_OBOE_H_
#define _AOWOO_OBOE_H_

#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif
	int32_t aowoo_open_oboe(int rate, int bit_depth, int chs, int frames);
	int32_t aowoo_pause_oboe(int state);

#ifdef __cplusplus
}
#endif

#endif
