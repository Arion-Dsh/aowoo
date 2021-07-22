// most code from
// https://www.alsa-project.org/alsa-doc/alsa-lib/_2test_2pcm_8c-example.html

// +build linux
#include <alsa/asoundlib.h>

const char*
aowoo_open_device(snd_pcm_t **handle,
					int channels,
					int rate,
					snd_pcm_uframes_t *period_size)
{

	const char *str;
	int err;
	snd_pcm_hw_params_t *params;

	// open pcm
	if (*handle == NULL) {
		char *device = "default";
		if ((err = snd_pcm_open(handle, device, SND_PCM_STREAM_PLAYBACK, 0)) < 0) {
			goto error;
		}
	}

	// set params
	err = snd_pcm_set_params(*handle, SND_PCM_FORMAT_FLOAT_LE, SND_PCM_ACCESS_RW_INTERLEAVED, channels, rate, 1, 2000);
	if (err < 0) {
		goto error;
	}

	// get params
	snd_pcm_uframes_t buffer_size;
	snd_pcm_get_params(*handle, &buffer_size, period_size);
	return str;

error:
	snd_pcm_hw_params_free(params);
	snd_pcm_close(*handle);
	str = snd_strerror(err);
	return str;
}
