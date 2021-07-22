// +build darwin

#include <AudioToolbox/AudioQueue.h>
#include <CoreAudio/CoreAudioTypes.h>
#include <CoreFoundation/CFRunLoop.h>

#include "_cgo_export.h"


#define AOWOO_NUM_BUFFERS 3
#define AOWOO_BUFFER_SIZE 4096

void aowoo_callback(void *id, AudioQueueRef queue, AudioQueueBufferRef buffer);


int aowoo_open(
        void *ctx,
        AudioQueueRef **queue,
        int rate,
        int chs)
{

    AudioQueueRef q;
    AudioStreamBasicDescription format;
    AudioQueueBufferRef buffers[3];

    format.mSampleRate       = rate;
    format.mFormatID         = kAudioFormatLinearPCM;
    format.mFormatFlags      = kAudioFormatFlagIsFloat;

    format.mBytesPerPacket      = chs * 4;
    format.mFramesPerPacket     = 1;
    format. mBytesPerFrame      = chs * 4;
    format.mChannelsPerFrame    = chs;
    format.mBitsPerChannel      = 8 * 4;


    AudioQueueNewOutput(&format, aowoo_callback, NULL, 0, kCFRunLoopCommonModes, 0, &q);
    *queue = &q;

    unsigned int i;
    for (i = 0; i < AOWOO_NUM_BUFFERS; i++)
    {
        AudioQueueAllocateBuffer(q, AOWOO_BUFFER_SIZE, &buffers[i]);

        buffers[i]->mAudioDataByteSize = AOWOO_BUFFER_SIZE;

        aowoo_callback(ctx, q, buffers[i]);
    }

    AudioQueueStart(q, NULL);

    return 0;


}

void aowoo_pause(AudioQueueRef *queue, char s){
    if (s == 1) {
        AudioQueuePause(*queue);
        return;
    }
    AudioQueueStart(*queue, NULL);
}

void aowoo_close(AudioQueueRef *queue)
{
        AudioQueueStop(*queue, false);
        AudioQueueDispose(*queue, false);
}


void aowoo_callback(void *ctx, AudioQueueRef queue, AudioQueueBufferRef buffer)
{
    float *casted_buffer = (float *)buffer->mAudioData;
    read_data(casted_buffer, AOWOO_BUFFER_SIZE/4, ctx);
    AudioQueueEnqueueBuffer(queue, buffer, 0, NULL);

}
